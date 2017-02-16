package parse

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	// MutilateAvg represent avg.
	MutilateAvg = "avg"
	// MutilateStd represent std.
	MutilateStd = "std"
	// MutilateMin represent min.
	MutilateMin = "min"
	// MutilatePercentile1st represent 1st latency percentile [us].
	MutilatePercentile1st = "percentile/1st"
	// MutilatePercentile5th represent 5th latency percentile [us].
	MutilatePercentile5th = "percentile/5th"
	// MutilatePercentile10th represent 10th latency percentile [us].
	MutilatePercentile10th = "percentile/10th"
	// MutilatePercentile90th represent 90th latency percentile [us].
	MutilatePercentile90th = "percentile/90th"
	// MutilatePercentile95th represent 95th latency percentile [us].
	MutilatePercentile95th = "percentile/95th"
	// MutilatePercentile99th represent 99th latency percentile [us].
	MutilatePercentile99th = "percentile/99th"
	// MutilateQPS represent qps.
	MutilateQPS = "qps"
)

// Results contains map of parsed mutilate results indexed by a name.
// TODO(bplotka): We should introduce here a typed struct instead of having dynamic map.
// We don't expect to change these fields any soon.
type Results struct {
	Raw               map[string]float64
	LatencyPercentile string
}

func newResults() Results {
	return Results{
		Raw: make(map[string]float64, 0),
	}
}

// File parse the file from given path and gather all metrics.
// NOTE: Public to allow use it without snap infrastructure.
func File(path string) (Results, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return newResults(), err
	}
	return Parse(file)
}

// Parse retrieves latency metrics from mutilate output. Following format is expected:
// #type       avg     std     min     5th    10th    90th    95th    99th
// read      109.6   231.8    17.4    49.4    55.9   137.2   216.1   916.0
func Parse(reader io.Reader) (Results, error) {
	metrics := newResults()
	scanner := bufio.NewScanner(reader)
	columnToMetric := map[int]string{}

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return newResults(), err
		}

		line := scanner.Text()
		if strings.HasPrefix(line, "#type") {
			m, err := parseHeader(line)
			if err != nil {
				return newResults(), err
			}

			columnToMetric = m

		} else if strings.HasPrefix(line, "read") {
			latencies, err := parseLatencies(line, columnToMetric)
			if err != nil {
				return newResults(), err
			}

			// This depends on the fact that the 'read...' line comes before the
			// custom latency and qps line. If other 'latencies' lines needs to be
			// parsed, the below should be a 'add to output map' rather than
			// overwriting.
			metrics.Raw = latencies

		} else if strings.HasPrefix(line, "Total QPS") {
			qps, err := parseQPS(line)
			if err != nil {
				return newResults(), err
			}

			metrics.Raw[MutilateQPS] = qps
		}
	}

	return metrics, nil
}

func parseHeader(line string) (map[int]string, error) {
	fields := strings.Fields(line)
	result := map[int]string{}

	if fields[0] != "#type" {
		return result, fmt.Errorf("Expect '%s' to start with '#type'", line)
	}

	labels := fields[1:]

	labelToMetric := map[string]string{
		"avg":  MutilateAvg,
		"std":  MutilateStd,
		"min":  MutilateMin,
		"1st":  MutilatePercentile1st,
		"5th":  MutilatePercentile5th,
		"10th": MutilatePercentile10th,
		"90th": MutilatePercentile90th,
		"95th": MutilatePercentile95th,
		"99th": MutilatePercentile99th,
		"qps":  MutilateQPS,
	}

	for index, label := range labels {
		metric, ok := labelToMetric[label]
		if !ok {
			return map[int]string{}, fmt.Errorf("No metric found for label '%s'", label)
		}

		result[index] = metric
	}

	return result, nil
}

// Parse line from Mutilate with read latencies. For example:
// "read       20.8    23.1    11.9    13.3    13.4    33.4    43.1    59.5"
// Returns a metrics map of {"avg": 20.8, "std": 23.1, ...}.
func parseLatencies(line string, columnToMetric map[int]string) (map[string]float64, error) {
	fields := strings.Fields(line)

	foundLatencies := (len(fields) - 1)
	expectedLatencies := len(columnToMetric)
	if foundLatencies != expectedLatencies {
		return map[string]float64{}, fmt.Errorf("Incorrect number of fields: expected %d but got %d", expectedLatencies, foundLatencies)
	}

	// Assume first field is the row type, for example 'read'.
	latencies := fields[1:]

	// Store latencies according to discovered column to metric mapping (from table header).
	result := map[string]float64{}

	for index, latency := range latencies {
		// Convert latency string to float64.
		value, err := strconv.ParseFloat(latency, 64)
		if err != nil {
			return map[string]float64{}, fmt.Errorf("'%s' latency value must be a float: %s", latency, err.Error())
		}

		key, ok := columnToMetric[index]
		if !ok {
			return map[string]float64{}, fmt.Errorf("No metric found for column index %d", index)
		}

		result[key] = value
	}

	return result, nil
}

// Parse the measured number of queries per second for latency measurement.
// For example: "Total QPS = 4993.1 (149793 / 30.0s)".
// Returns value. For example 4993.1
func parseQPS(line string) (float64, error) {
	var (
		qps      float64
		count    int
		duration float64
	)

	const fields = 3

	if n, err := fmt.Sscanf(line, "Total QPS = %f (%d / %fs)", &qps, &count, &duration); err != nil {
		if n != fields {
			return 0.0, fmt.Errorf("Incorrect number of fields: expected %d but got %d", fields, n)
		}

		return 0.0, err
	}

	return qps, nil
}
