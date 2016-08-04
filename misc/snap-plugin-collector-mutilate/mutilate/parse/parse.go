package parse

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	// MutilateAvg represent avg.
	MutilateAvg = "avg"
	// MutilateStd represent std.
	MutilateStd = "std"
	// MutilateMin represent min.
	MutilateMin = "min"
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
	// MutilatePercentileCustom represent custom latency percentile [us].
	MutilatePercentileCustom = "percentile/custom"
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

// File parse the file from given path and gather all metrics
// including (custom percentile).
// NOTE: Public to allow use it without snap infrastructure.
func File(path string) (Results, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return newResults(), err
	}
	return Parse(file)
}

// OpenedFile parse the file from given file handle and gather all metrics
// including (custom percentile). It leaves the responsibilities of this handler to the caller.
// NOTE: Public to allow use it without snap infrastructure.
func OpenedFile(file *os.File) (Results, error) {
	return Parse(file)
}

// Parse retrieves latency metrics from mutilate output. Following format is expected:
// #type       avg     std     min     5th    10th    90th    95th    99th
// read      109.6   231.8    17.4    49.4    55.9   137.2   216.1   916.0
func Parse(file io.Reader) (Results, error) {
	metrics := newResults()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return newResults(), err
		}

		line := scanner.Text()
		if strings.HasPrefix(line, "read") {
			latencies, err := parseReadLatencies(line)
			if err != nil {
				return newResults(), err
			}

			// This depends on the fact that the 'read...' line comes before the
			// custom latency and qps line. If other 'latencies' lines needs to be
			// parsed, the below should be a 'add to output map' rather than
			// overwriting.
			metrics.Raw = latencies

		} else if strings.HasPrefix(line, "Swan latency for percentile") {
			percentile, latency, err := parseCustomPercentileLatency(line)
			if err != nil {
				return newResults(), err
			}

			metrics.Raw[MutilatePercentileCustom] = latency
			metrics.LatencyPercentile = percentile

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

// Parse line from Mutilate with read latencies. For example:
// "read       20.8    23.1    11.9    13.3    13.4    33.4    43.1    59.5"
// Returns a metrics map of {"avg": 20.8, "std": 23.1, ...}.
func parseReadLatencies(line string) (map[string]float64, error) {
	var (
		avg float64
		std float64
		min float64
		p5  float64
		p10 float64
		p90 float64
		p95 float64
		p99 float64
	)

	const fields = 8

	if n, err := fmt.Sscanf(line, "read %f %f %f %f %f %f %f %f", &avg, &std, &min, &p5, &p10, &p90, &p95, &p99); err != nil {
		if n != fields {
			return nil, fmt.Errorf("Incorrect number of fields: expected %d but got %d", fields, n)
		}

		return nil, err
	}

	return map[string]float64{
		MutilateAvg:            avg,
		MutilateStd:            std,
		MutilateMin:            min,
		MutilatePercentile5th:  p5,
		MutilatePercentile10th: p10,
		MutilatePercentile90th: p90,
		MutilatePercentile95th: p95,
		MutilatePercentile99th: p99,
	}, nil
}

// Mutilate in the Swan repo has been patched to provide a custom percentile latency measurement.
// For example: "Swan latency for percentile 99.999000: 1777.887805"
// Returns a pair of resulted percentile and value. For example ('99.99900', 1777.887805).
func parseCustomPercentileLatency(line string) (string, float64, error) {
	var (
		percentile float64
		latency    float64
	)

	const fields = 2

	// TODO(bplotka): We might use Regexp here instead of sscanf to scan %s for percentile.
	if n, err := fmt.Sscanf(line, "Swan latency for percentile %f: %f", &percentile, &latency); err != nil {
		if n != fields {
			return "", 0.0, fmt.Errorf("Incorrect number of fields: expected %d but got %d", fields, n)
		}

		return "", 0.0, err
	}

	return fmt.Sprintf("%f", percentile), latency, nil
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
