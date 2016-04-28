package mutilate

import (
	"bufio"
	"bytes"
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
	output = append(output, scanMutilateStdoutRows(scanner, baseTime)...)

	return output, nil
}

func scanMutilateStdoutRows(scanner *bufio.Scanner, eventTime time.Time) []Metric {
	var output []Metric
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "read") {
			output = append(output, parseDefaultMutilateStdoutRow(line, eventTime)...)
		} else if strings.Contains(line, "Swan latency for percentile") {
			output = append(output, parseSwanMutilateStdoutRow(line, eventTime))
		}

	}

	return output
}

// parseDefaultMutilateStdoutRow takes row on input; first column is ignored as it is a row description,
// not a metric
// example row: read 20.8 23.1 11.9 13.3 13.4 33.4 43.1 59.5
func parseDefaultMutilateStdoutRow(line string, eventTime time.Time) []Metric {
	var output []Metric
	fields := strings.Fields(line)
	metricNames := getDefaultMetricsNames()
	for index, value := range fields {
		if index == 0 {
			continue
		}
		output = append(output,
			createMetricsFromDefaultStdout(value, metricNames[index], eventTime))
	}

	return output
}

func getDefaultMetricsNames() []string {
	var names []string
	names = append(names, "", "avg", "std", "min", "percentile/5th", "percentile/10th",
		"percentile/90th", "percentile/95th", "percentile/99th")

	return names

}

func createMetricsFromDefaultStdout(value string, name string, eventTime time.Time) Metric {
	floatValue, _ := strconv.ParseFloat(value, 64)
	metric := Metric{name, floatValue, eventTime}
	return metric
}

// parseSwanMutilateStdoutRow takes a row containing custom percentile data on input
// example row: Swan latency for percentile 99.999000: 1777.887805
func parseSwanMutilateStdoutRow(line string, eventTime time.Time) Metric {
	lineFields := strings.Split(line, ":")
	floatValue, _ := strconv.ParseFloat(strings.TrimSpace(lineFields[1]), 64)
	name := getMetricNameFromSwanRowDescription(lineFields[0])
	output := Metric{name, floatValue, eventTime}

	return output
}

func getMetricNameFromSwanRowDescription(description string) string {
	words := strings.Split(description, " ")
	percentileName := words[len(words)-1]
	var buffer bytes.Buffer
	buffer.WriteString("percentile/")
	buffer.WriteString(strings.Trim(percentileName, "0"))
	buffer.WriteString("th")

	return buffer.String()
}
