package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Results has a map of results indexed by a name.
type Results struct {
	Raw map[string]uint64
}

func newResults() Results {
	return Results{
		Raw: make(map[string]uint64, 0),
	}
}

// File parse the file from given path.
func File(path string) (Results, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return newResults(), err
	}
	return Parse(file)
}

// Parse retrieves metrics from specjbb output represented as:
// 262s: Performance info:
// Transaction,    Success,    Partial,     Failed,   Receipts, AvgBarcode,
// Overall,         122034,          0,          0,     115656,      42.09,
// Response times:
// Request,          Success,    Partial,     Failed,   SkipFail,     Probes,    Samples,      min,      p50,      p90,      p95,      p99,      max,
// TotalPurchase,     128453,          0,          0,          0,        127,     171506,  3800000,  6600000,  7400000,  7400000,  7700000,  8000000,
func Parse(reader io.Reader) (Results, error) {
	metrics := newResults()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return newResults(), err
		}

		line := strings.Join(strings.Fields(scanner.Text()), "")
		if strings.HasPrefix(line, "TotalPurchase,") {
			latencies, err := parseReadLatencies(line)
			if err != nil {
				return newResults(), err
			}
			metrics.Raw = latencies
		}
	}
	return metrics, nil
}

// Parse line from specjbb with latencies of TotalPurchase. For example:
// TotalPurchase,     128453,          0,          0,          0,        127,     171506,  3800000,  6600000,  7400000,  7400000,  7700000,  8000000,
// Returns a map of {"Success": 128453, "Partial": 0, ...}.
func parseReadLatencies(line string) (map[string]uint64, error) {
	fmt.Println(line)
	var (
		success  uint64
		partial  uint64
		failed   uint64
		skipFail uint64
		probes   uint64
		samples  uint64
		min      uint64
		p50      uint64
		p90      uint64
		p95      uint64
		p99      uint64
		max      uint64
	)

	const fields = 12

	if numberOfItems, err := fmt.Sscanf(line, "TotalPurchase,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,",
		&success, &partial, &failed, &skipFail, &probes, &samples, &min, &p50, &p90, &p95, &p99, &max); err != nil {
		if numberOfItems != fields {
			return nil, fmt.Errorf("Incorrect number of fields: expected %d but got %d", fields, numberOfItems)
		}

		return nil, err
	}

	return map[string]uint64{
		"Success":  success,
		"Partial":  partial,
		"Failed":   failed,
		"SkipFail": skipFail,
		"Probes":   probes,
		"Samples":  samples,
		"min":      min,
		"p50":      p50,
		"p90":      p90,
		"p95":      p95,
		"p99":      p99,
		"max":      max,
	}, nil
}
