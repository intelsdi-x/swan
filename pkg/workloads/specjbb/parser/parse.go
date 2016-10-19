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

// FileWithRawFileName parse the file with raw file name from given path.
func FileWithRawFileName(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", err
	}
	return ParseRawFileName(file)
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

// FileWithHBIRRT parse the file with HBIR output from given path.
func FileWithHBIRRT(path string) (int, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return 0, err
	}
	return ParseHBIRRT(file)
}

// ParseRawFileName retrieves name of raw file from specjbb output represented as:
// 6s: Binary log file is /home/vagrant/go/src/github.com/intelsdi-x/swan/workloads/web_serving/specjbb/specjbb2015-D-20160921-00002.data.gz
func ParseRawFileName(reader io.Reader) (string, error) {
	var rawFileName string
	var time int
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		// Remove whitespaces, as SPECjbb generates random number of spaces to create a good-looking table.
		// To parse output we need a constant form of it.
		line := strings.Join(strings.Fields(scanner.Text()), "")
		if strings.Contains(line, "Binarylogfileis") {
			const fields = 2
			if numberOfItems, err := fmt.Sscanf(line, "%ds:Binarylogfileis%s", &time, &rawFileName); err != nil {
				if numberOfItems != fields {
					return "", fmt.Errorf("Incorrect number of fields: expected %d but got %d", fields, numberOfItems)
				}
				return "", err
			}
			return rawFileName, nil
		}
	}
	return "", fmt.Errorf("Raw file name not found")
}

// ParseHBIRRT retrieves geo mean of critical jops from specjbb output represented as:
// RUN RESULT: hbIR (max attempted) = 12000, hbIR (settled) = 12000, max-jOPS = 11640, critical-jOPS = 2684
func ParseHBIRRT(reader io.Reader) (int, error) {
	var hbir int
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, err
		}
		// Remove whitespaces, as SPECjbb generates random number of spaces to create a good-looking table.
		// To parse output we need a constant form of it.
		line := strings.Join(strings.Fields(scanner.Text()), "")
		if strings.Contains(line, "RUNRESULT:") && strings.Contains(line, "critical-jOPS") {
			// We can't create universal regex for this line as sometimes values are int and sometimes string.
			// Instead we have to divide it by comas and take last element and then extract its value.
			// Split output to have results in separate fields.
			lineSplitted := strings.Split(line, ",")
			// Last element in a line is always a critical jops.
			lastElement := lineSplitted[len(lineSplitted)-1]
			const fields = 1
			if numberOfItems, err := fmt.Sscanf(lastElement, "critical-jOPS=%d", &hbir); err != nil {
				if numberOfItems != fields {
					return 0, fmt.Errorf("Incorrect number of fields: expected %d but got %d", fields, numberOfItems)
				}
				return 0, err
			}
			return hbir, nil
		}
	}
	return 0, fmt.Errorf("Run result not found")
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

		// Remove whitespaces, as SPECjbb generates random number of spaces to create a good-looking table.
		// To parse output we need a constant form of it.
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
		"Success":         success,
		"Partial":         partial,
		"Failed":          failed,
		"SkipFail":        skipFail,
		"Probes":          probes,
		"Samples":         samples,
		"min":             min,
		"percentile/50th": p50,
		"percentile/90th": p90,
		"percentile/95th": p95,
		"percentile/99th": p99,
		"max":             max,
	}, nil
}
