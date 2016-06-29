package parse

import (
	"bufio"
	"fmt"
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
	// MutilateQPS represent qps.
	MutilateQPS = "qps"
)

// Metrics is a type alias for a float map indexed by a name.
// TODO(bplotka): We should introduce here a typed struct instead of having dynamic map.
// We don't expect to change these fields any soon.
type Metrics map[string]float64

// GenerateCustomPercentileKey returns a key of custom percentile field in Metrics.
// TODO(bplotka): This needs to be erased when typed struct will replace dynamic map.
func GenerateCustomPercentileKey(customPercentile float64) string {
	return fmt.Sprintf("percentile/%2.3fth/custom", customPercentile)
}

// File parse the file from given path and gather all metrics
// including (custom percentile).
// NOTE: Public to allow use it without snap infrastructure.
func File(path string) (Metrics, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return Metrics{}, err
	}
	return parse(file)
}

// OpenedFile parse the file from given file handle and gather all metrics
// including (custom percentile). It leaves the responsibilities of this handler to the caller.
// NOTE: Public to allow use it without snap infrastructure.
func OpenedFile(file *os.File) (Metrics, error) {
	return parse(file)
}

func parse(file *os.File) (Metrics, error) {
	metrics := Metrics{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return Metrics{}, err
		}

		line := scanner.Text()
		if strings.HasPrefix(line, "read") {
			latencies, err := parseReadLatencies(line)
			if err != nil {
				return Metrics{}, err
			}

			// This depends on the fact that the 'read...' line comes before the
			// custom latency and qps line. If other 'latencies' lines needs to be
			// parsed, the below should be a 'add to output map' rather than
			// overwriting.
			metrics = latencies

		} else if strings.HasPrefix(line, "Swan latency for percentile") {
			name, latency, err := parseCustomPercentileLatency(line)
			if err != nil {
				return Metrics{}, err
			}

			metrics[name] = latency

		} else if strings.HasPrefix(line, "Total QPS") {
			name, qps, err := parseQPS(line)
			if err != nil {
				return Metrics{}, err
			}

			metrics[name] = qps
		}
	}

	return metrics, nil
}

// Parse line from Mutilate with read latencies. For example:
// "read       20.8    23.1    11.9    13.3    13.4    33.4    43.1    59.5"
// Returns a metrics map of {"avg": 20.8, "std": 23.1, ...}.
func parseReadLatencies(line string) (Metrics, error) {
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
			return Metrics{}, fmt.Errorf("Incorrect number of fields: expected %d but got %d", fields, n)
		}

		return Metrics{}, err
	}

	return Metrics{
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
// Returns a pair of metric name and value. For example ('percentile/99.999th/custom', 1777.887805).
func parseCustomPercentileLatency(line string) (string, float64, error) {
	var (
		percentile float64
		latency    float64
	)

	const fields = 2

	if n, err := fmt.Sscanf(line, "Swan latency for percentile %f: %f", &percentile, &latency); err != nil {
		if n != fields {
			return "", 0.0, fmt.Errorf("Incorrect number of fields: expected %d but got %d", fields, n)
		}

		return "", 0.0, err
	}

	return GenerateCustomPercentileKey(percentile), latency, nil
}

// Parse the measured number of queries per second for latency measurement.
// For example: "Total QPS = 4993.1 (149793 / 30.0s)".
// Returns a pair of metric name and value. For example ('qps', 4993.1).
func parseQPS(line string) (string, float64, error) {
	var (
		qps      float64
		count    int
		duration float64
	)

	const fields = 3

	if n, err := fmt.Sscanf(line, "Total QPS = %f (%d / %fs)", &qps, &count, &duration); err != nil {
		if n != fields {
			return "", 0.0, fmt.Errorf("Incorrect number of fields: expected %d but got %d", fields, n)
		}

		return "", 0.0, err
	}

	return MutilateQPS, qps, nil
}
