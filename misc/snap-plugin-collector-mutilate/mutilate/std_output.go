package mutilate

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/exp/inotify"
	"os"
	"strconv"
	"strings"
	"time"
)

// Metric represents sigle metric retrieved from mutilate standard output
type Metric struct {
	name  string
	value float64
	time  time.Time
}

func parseMutilateStdout(event inotify.Event, baseTime time.Time) ([]Metric, error) {
	var output []Metric
	csvFile, readError := os.Open(event.Name)
	defer csvFile.Close()
	if readError != nil {
		return output, readError
	}
	scanner := bufio.NewScanner(csvFile)
	metrics, scanningError := scanMutilateStdoutRows(scanner, baseTime)
	if scanningError != nil {
		return output, scanningError
	}
	output = append(output, metrics...)

	return output, nil
}

func scanMutilateStdoutRows(scanner *bufio.Scanner, eventTime time.Time) ([]Metric, error) {
	var output, defaultRow []Metric
	var swanRow Metric
	var defaultRowError, swanRowError error
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "read") {
			defaultRow, defaultRowError = parseDefaultMutilateStdoutRow(line, eventTime)
			if defaultRowError != nil {
				return output, defaultRowError
			}

		} else if strings.Contains(line, "Swan latency for percentile") {
			swanRow, swanRowError = parseSwanMutilateStdoutRow(line, eventTime)
			if swanRowError != nil {
				return output, swanRowError
			}
		}

	}
	if len(defaultRow) == 0 {
		return output, errors.New("No default mutilate statistics found")
	}
	if swanRow.name == "" && swanRow.value == 0.0 {
		return output, errors.New("No swan-specific statistics found")
	}
	output = append(output, defaultRow...)
	output = append(output, swanRow)

	return output, nil
}

// parseDefaultMutilateStdoutRow takes row on input; first column is ignored as it is a row description,
// not a metric.
// example row: read 20.8 23.1 11.9 13.3 13.4 33.4 43.1 59.5
func parseDefaultMutilateStdoutRow(line string, eventTime time.Time) ([]Metric, error) {
	var output []Metric
	fields := strings.Fields(line)
	if colCount := len(fields); colCount != 9 {
		return output, errors.New(fmt.Sprintf("Incorrect column count (got: %d, expected: 9) in QPS read row", colCount))
	}
	metricNames := getDefaultMetricsNames()
	for index, value := range fields {
		if index == 0 {
			continue
		}
		metric, metricError := createMetricsFromDefaultStdout(value, metricNames[index], eventTime)
		if metricError != nil {
			return output, errors.New(fmt.Sprintf("Non-numeric value in read row (value: \"%s\", column: %d)", metricError.Error(), index+1))
		}
		output = append(output, metric)
	}

	return output, nil
}

func getDefaultMetricsNames() []string {
	var names []string
	names = append(names, "", "avg", "std", "min", "percentile/5th", "percentile/10th",
		"percentile/90th", "percentile/95th", "percentile/99th")

	return names

}

func createMetricsFromDefaultStdout(value string, name string, eventTime time.Time) (Metric, error) {
	floatValue, floatError := strconv.ParseFloat(value, 64)
	if floatError != nil {
		return Metric{}, errors.New(value)
	}
	metric := Metric{name, floatValue, eventTime}
	return metric, nil
}

// parseSwanMutilateStdoutRow takes a row containing custom percentile data on input.
// example row: Swan latency for percentile 99.999000: 1777.887805
func parseSwanMutilateStdoutRow(line string, eventTime time.Time) (Metric, error) {
	lineFields := strings.Split(line, ":")
	if len(lineFields) != 2 {
		return Metric{}, errors.New("Swan-specific row malformed")
	}
	floatValue, floatParsingError := strconv.ParseFloat(strings.TrimSpace(lineFields[1]), 64)
	if floatParsingError != nil {
		return Metric{}, errors.New("Swan-specific row is missing metric value")
	}
	name, nameError := getMetricNameFromSwanRowDescription(lineFields[0])
	if nameError != nil {
		return Metric{}, nameError
	}
	output := Metric{name, floatValue, eventTime}

	return output, nil
}

func getMetricNameFromSwanRowDescription(description string) (string, error) {
	words := strings.Split(description, " ")
	percentileName := strings.Trim(words[len(words)-1], "0")
	if _, floatError := strconv.ParseFloat(percentileName, 64); floatError != nil {
		return "", errors.New("Swan-specific row is missing percentile value")
	}
	var buffer bytes.Buffer
	buffer.WriteString("percentile/")
	buffer.WriteString(percentileName)
	buffer.WriteString("th")

	return buffer.String(), nil
}
