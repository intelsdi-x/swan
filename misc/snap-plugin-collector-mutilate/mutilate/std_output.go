package mutilate

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Metric represents sigle metric retrieved from mutilate standard output
type metric struct {
	name  string
	value float64
}

func (m metric) isMetricEmpty() bool {
	return m.name == "" && m.value == 0.0
}

func parseMutilateStdout(filePath string) ([]metric, error) {
	var output []metric
	csvFile, readError := os.Open(filePath)
	defer csvFile.Close()
	if readError != nil {
		return output, readError
	}
	scanner := bufio.NewScanner(csvFile)
	metrics, scanningError := scanMutilateStdoutRows(scanner)
	if scanningError != nil {
		return output, scanningError
	}
	output = append(output, metrics...)

	return output, nil
}

func scanMutilateStdoutRows(scanner *bufio.Scanner) ([]metric, error) {
	var output, defaultMetrics []metric
	var swanMetric metric
	var err error
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return output, err
		}
		row := scanner.Text()
		if strings.Contains(row, "read") {
			defaultMetrics, err = parseDefaultMutilateStdoutRow(row)
			if err != nil {
				return output, err
			}

		} else if strings.Contains(row, "Swan latency for percentile") {
			swanMetric, err = parseSwanMutilateStdoutRow(row)
			if err != nil {
				return output, err
			}
		}

	}
	if defaultMetrics == nil {
		return output, errors.New("No default mutilate statistics found")
	}
	if swanMetric.isMetricEmpty() {
		return output, errors.New("No swan-specific statistics found")
	}
	output = append(output, defaultMetrics...)
	output = append(output, swanMetric)

	return output, nil
}

// parseDefaultMutilateStdoutRow takes row on input; first column is ignored as it is a row description,
// not a metric.
// example row: read 20.8 23.1 11.9 13.3 13.4 33.4 43.1 59.5
func parseDefaultMutilateStdoutRow(line string) ([]metric, error) {
	var output []metric
	fields := strings.Fields(line)
	if colCount := len(fields); colCount != 9 {
		return output, fmt.Errorf("Incorrect column count (got: %d, expected: 9) in "+
			"QPS read row", colCount)
	}
	metricNames := [...]string{"", "avg", "std", "min", "percentile/5th",
		"percentile/10th", "percentile/90th", "percentile/95th", "percentile/99th"}
	for index, value := range fields {
		if index == 0 {
			continue
		}
		metric, err := newMetricFrom(value, metricNames[index])
		if err != nil {
			return output, fmt.Errorf("Non-numeric value in read row (value: \"%s\","+
				" column: %d)", err.Error(), index+1)
		}
		output = append(output, metric)
	}

	return output, nil
}

func newMetricFrom(value string, name string) (metric, error) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return metric{}, errors.New(value)
	}
	defaultMetric := metric{name, floatValue}

	return defaultMetric, nil
}

// parseSwanMutilateStdoutRow takes a row containing custom percentile data on input.
// example row: Swan latency for percentile 99.999000: 1777.887805
func parseSwanMutilateStdoutRow(line string) (metric, error) {
	lineFields := strings.Split(line, ":")
	if len(lineFields) != 2 {
		return metric{}, errors.New("Swan-specific row malformed")
	}
	floatValue, floatParsingError := strconv.ParseFloat(
		strings.TrimSpace(lineFields[1]), 64)
	if floatParsingError != nil {
		return metric{}, errors.New("Swan-specific row is missing metric value")
	}
	name, nameError := getMetricNameFromSwanRowDescription(lineFields[0])
	if nameError != nil {
		return metric{}, nameError
	}
	output := metric{name, floatValue}

	return output, nil
}

func getMetricNameFromSwanRowDescription(description string) (string, error) {
	words := strings.Split(description, " ")
	percentileName := strings.Trim(words[len(words)-1], "0")
	if _, err := strconv.ParseFloat(percentileName, 64); err != nil {
		return "", errors.New("Swan-specific row is missing percentile value")
	}
	var buffer bytes.Buffer
	buffer.WriteString("percentile/")
	buffer.WriteString(percentileName)
	buffer.WriteString("th")

	return buffer.String(), nil
}
