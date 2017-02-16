package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	// SuccessKey is a key for success metric in SPECjbb controller output.
	SuccessKey = "Success"
	// PartialKey is a key for partial metric in SPECjbb controller output.
	PartialKey = "Partial"
	// FailedKey is a key for failed metric in SPECjbb controller output.
	FailedKey = "Failed"
	// SkipFailKey is a key for skipFail metric in SPECjbb controller output.
	SkipFailKey = "SkipFail"
	// ProbesKey is a key for probes metric in SPECjbb controller output.
	ProbesKey = "Probes"
	// SamplesKey is a key for samples metric in SPECjbb controller output.
	SamplesKey = "Samples"
	// MinKey is a key for min latency metric in SPECjbb controller output.
	MinKey = "min"
	// Percentile50Key is a key for 50th percentile metric in SPECjbb controller output.
	Percentile50Key = "percentile/50th"
	// Percentile90Key is a key for 90th percentile metric in SPECjbb controller output.
	Percentile90Key = "percentile/90th"
	// Percentile95Key is a key for 95th percentile metric in SPECjbb controller output.
	Percentile95Key = "percentile/95th"
	// Percentile99Key is a key for 99th percentile metric metric in SPECjbb controller output.
	Percentile99Key = "percentile/99th"
	// MaxKey is a key for max latency metric in SPECjbb controller output.
	MaxKey = "max"
	// QPSKey is a key for processed requests metric in SPECjbb controller output.
	QPSKey = "qps"
	// IssuedRequestsKey is a key for actual injection rate metric in SPECjbb controller output.
	IssuedRequestsKey = "issued_requests"
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
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		// Remove whitespaces, as SPECjbb generates random number of spaces to create a good-looking table.
		// To parse output we need a constant form of it.
		// <Mon Nov 14 14:53:39 CET 2016> org.spec.jbb.controller: Controller start
		// Binary log file is /tmp/specjbb2015-D-20161114-00007.data.gz
		// 1s:
		// 1s: <Mon Nov 14 14:53:39 CET 2016> org.spec.jbb.controller: Binary log file is /tmp/specjbb2015-D-20161114-00007.data.gz
		line := strings.Join(strings.Fields(scanner.Text()), "")
		regex := regexp.MustCompile("Binarylogfileis([/a-zA-Z-_.0-9]+)")
		if regex.MatchString(line) {
			submatch := regex.FindStringSubmatch(line)
			// Submatch should have length = 2, First is matched string, the second is value of a group (name of a file).
			if len(submatch) < 2 {
				return "", errors.New("Raw file name not found")
			}
			rawFileName := submatch[len(submatch)-1]
			return rawFileName, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", errors.Wrap(err, "could read from io.reader")
	}
	return "", errors.New("Raw file name not found")
}

// ParseHBIRRT retrieves geo mean of critical jops from specjbb output represented as:
// RUN RESULT: hbIR (max attempted) = 12000, hbIR (settled) = 12000, max-jOPS = 11640, critical-jOPS = 2684
func ParseHBIRRT(reader io.Reader) (int, error) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		// Remove whitespaces, as SPECjbb generates random number of spaces to create a good-looking table.
		// To parse output we need a constant form of it.
		line := strings.Join(strings.Fields(scanner.Text()), "")
		regex := regexp.MustCompile("RUNRESULT:[()a-zA-Z,0-9=-]+critical-jOPS=([0-9]+)")
		if regex.MatchString(line) {
			submatch := regex.FindStringSubmatch(line)
			// Submatch should have length = 2, First is matched string, the second is value of a group (value of jops).
			if len(submatch) < 2 {
				return 0, errors.New("Run result not found, cannot determine critical-jops")
			}
			hbir, err := strconv.Atoi(submatch[len(submatch)-1])
			if err != nil {
				return 0, errors.New("Bad value type found for critical-jops")
			}
			return hbir, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, errors.Wrap(err, "cannot parse HBIRRT: could not read from io.reader")
	}
	return 0, errors.New("Run result not found, cannot determine critical-jops")
}

// ParseLatencies retrieves metrics from specjbb output represented as:
// 55s: ( 0%) ......|................?............. (rIR:aIR:PR = 4000:4007:4007) (tPR = 60729) [OK]
// 262s: Performance info:
// Transaction,    Success,    Partial,     Failed,   Receipts, AvgBarcode,
// Overall,         122034,          0,          0,     115656,      42.09,
// Response times:
// Request,          Success,    Partial,     Failed,   SkipFail,     Probes,    Samples,      min,      p50,      p90,      p95,      p99,      max,
// TotalPurchase,     128453,          0,          0,          0,        127,     171506,  3800000,  6600000,  7400000,  7400000,  7700000,  8000000,
func ParseLatencies(reader io.Reader) (Results, error) {
	metrics := newResults()
	scanner := bufio.NewScanner(reader)
	metricsRaw := make(map[string]uint64, 0)
	// Regex for line with actual injection rate and processed requests.
	// 55s: ( 0%) ......|................?............. (rIR:aIR:PR = 4000:4007:4007) (tPR = 60729) [OK]
	rLocal := regexp.MustCompile("[0-9]+s:[ ()0-9%.|?]+rIR:aIR:PR[ =]+([0-9]+):([0-9]+):([0-9]+)")
	// Try to match two types logs below:
	// <Wed Nov 09 18:58:39 UTC 2016> org.spec.jbb.controller: PRESET: IR = 500 finished, steady status = [OK] (rIR:aIR:PR = 500:500:500) (tPR = 7214)
	// or
	// <Fri Dec 16 16:06:35 CET 2016> org.spec.jbb.controller: PRESET: IR = 4000 finished, settle status = [PR is under limit] (rIR:aIR:PR = 4000:3960:3350) (tPR = 48530)
	// (rIR:aIR:PR = 4000:3960:3350) (tPR = 48530) [PR is under limit]
	rRemote := regexp.MustCompile("[<a-zA-Z:0-9]+PRESET:[a-zA-Z=0-9]+finished,(steady|settle)status=\\[[a-zA-Z]+\\][()]rIR:aIR:PR=([0-9]+):([0-9]+):([0-9]+)")
	for scanner.Scan() {
		// Remove whitespaces, as SPECjbb generates random number of spaces to create a good-looking table.
		// To parse output we need a constant form of it.
		line := strings.Join(strings.Fields(scanner.Text()), "")
		var submatch []string
		if rLocal.MatchString(line) {
			submatch = rLocal.FindStringSubmatch(line)
		} else if rRemote.MatchString(line) {
			submatch = rRemote.FindStringSubmatch(line)
		} else if strings.HasPrefix(line, "TotalPurchase,") {
			latencies, err := parseTotalPurchaseLatencies(line)
			if err != nil {
				return newResults(), err
			}
			metricsRaw = mapCopy(latencies, metricsRaw)
		}
		if len(submatch) > 0 {
			issuedRequests, processedRequests, err := parseRequests(submatch)
			if err != nil {
				return newResults(), err
			}
			metricsRaw[QPSKey] = processedRequests
			metricsRaw[IssuedRequestsKey] = issuedRequests
		}
	}
	if err := scanner.Err(); err != nil {
		return newResults(), errors.Wrap(err, "cannot parse latencies: cannot read from io.reader")
	}
	metrics.Raw = metricsRaw

	_, ok := metrics.Raw[QPSKey]
	if !ok {
		return newResults(), errors.New("cannot find processed requests value (PR) in SPECjbb controller output")
	}
	_, ok = metrics.Raw[IssuedRequestsKey]
	if !ok {
		return newResults(), errors.New("cannot find issued requests value (aIR) in SPECjbb controller output")
	}
	return metrics, nil
}

// parseRequests returns two values:
// - issued requests - number of requests issued by transaction injector to backend
// - processed requests - number of requests processed by backend
// by parsing below line from the controller output:
// 55s: ( 0%) ......|................?............. (rIR:aIR:PR = 4000:4007:4007) (tPR = 60729) [OK]
func parseRequests(submatch []string) (uint64, uint64, error) {
	// Returned submatch should have 4 fields: matched string, requested IR, actual IR and processed requests.
	if len(submatch) >= 4 {
		// We use actual injection rate and processed requests (two last values in a slice).
		processedRequests, err := strconv.ParseUint(submatch[len(submatch)-1], 10, 64)
		if err != nil {
			return 0, 0, errors.New("invalid type of processed requests value (PR) in SPECjbb controller output")
		}
		issuedRequests, err := strconv.ParseUint(submatch[len(submatch)-2], 10, 64)
		if err != nil {
			return 0, 0, errors.New("invalid type of issued requests value (aIR) in SPECjbb controller output")
		}
		return issuedRequests, processedRequests, nil
	}

	return 0, 0, errors.New("cannot find processed requests value in SPECjbb controller output")

}

func mapCopy(source, destination map[string]uint64) map[string]uint64 {
	for k, v := range source {
		destination[k] = v
	}
	return destination
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
			return nil, errors.Errorf("Incorrect number of fields: expected %d but got %d", fields, numberOfItems)
		}

		return nil, err
	}

	return map[string]uint64{
		SuccessKey:      success,
		PartialKey:      partial,
		FailedKey:       failed,
		SkipFailKey:     skipFail,
		ProbesKey:       probes,
		SamplesKey:      samples,
		MinKey:          min,
		Percentile50Key: p50,
		Percentile90Key: p90,
		Percentile95Key: p95,
		Percentile99Key: p99,
		MaxKey:          max,
	}, nil
}
