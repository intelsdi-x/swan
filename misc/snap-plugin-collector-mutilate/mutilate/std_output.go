package mutilate

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Metrics is a type alias for a float map indexed by a name.
type Metrics map[string]float64

func parse(path string) (Metrics, error) {
	csvFile, err := os.Open(path)
	defer csvFile.Close()
	if err != nil {
		return Metrics{}, err
	}

	metrics := Metrics{}
	scanner := bufio.NewScanner(csvFile)
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
			name, latency, err := parseCustomPercentile(line)
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
		"avg":             avg,
		"std":             std,
		"min":             min,
		"percentile/5th":  p5,
		"percentile/10th": p10,
		"percentile/90th": p90,
		"percentile/95th": p95,
		"percentile/99th": p99,
	}, nil
}

func parseCustomPercentile(line string) (string, float64, error) {
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

	return fmt.Sprintf("percentile/%2.3fth/custom", percentile), latency, nil
}

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

	return "qps/total", qps, nil
}
