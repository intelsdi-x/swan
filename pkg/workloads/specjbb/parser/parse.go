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

// FileWithLatencies parse the file with load output from given path.
func FileWithLatencies(path string) (Results, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return newResults(), err
	}
	return ParseLatencies(file)
}

// FileWithHBIR parse the file with HBIR output from given path.
func FileWithHBIR(path string) (int, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return 0, err
	}
	return ParseHBIR(file)
}

// ParseHBIR retrieves HBIR from specjbb output represented as:
//    559s: high-bound for max-jOPS is measured to be 6725
func ParseHBIR(reader io.Reader) (int, error) {
	var hbir, timeOfCalculation int
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, err
		}

		// Remove whitespaces because SPECjbb adds different number of them to generate 'table'.
		// To check whether a line contains specific string we must have constant format of it.
		line := strings.Join(strings.Fields(scanner.Text()), "")
		if strings.Contains(line, "high-boundformax-jOPSismeasuredtobe") {
			const fields = 2
			if numberOfItems, err := fmt.Sscanf(line, "%ds:high-boundformax-jOPSismeasuredtobe%d", &timeOfCalculation, &hbir); err != nil {
				if numberOfItems != fields {
					return 0, fmt.Errorf("Incorrect number of fields: expected %d but got %d", fields, numberOfItems)
				}
				return 0, err
			}
		}
	}
	return hbir, nil
}

// ParseLatencies retrieves metrics from specjbb output represented as:
// 262s: Performance info:
// Transaction,    Success,    Partial,     Failed,   Receipts, AvgBarcode,
// Overall,         122034,          0,          0,     115656,      42.09,
// Response times:
// Request,          Success,    Partial,     Failed,   SkipFail,     Probes,    Samples,      min,      p50,      p90,      p95,      p99,      max,
// TotalPurchase,     128453,          0,          0,          0,        127,     171506,  3800000,  6600000,  7400000,  7400000,  7700000,  8000000,
func ParseLatencies(reader io.Reader) (Results, error) {
	metrics := newResults()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return newResults(), err
		}

		line := strings.Join(strings.Fields(scanner.Text()), "")
		if strings.HasPrefix(line, "TotalPurchase,") {
			latencies, err := parseTotalPurchaseLatencies(line)
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
func parseTotalPurchaseLatencies(line string) (map[string]uint64, error) {
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
