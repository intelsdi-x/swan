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

type MutilateMetric struct {
	name  string
	value float64
	time  time.Time
}

func parse_mutilate_stdout(event inotify.Event, baseTime time.Time) ([]MutilateMetric, error) {
	var output []MutilateMetric
	csvFile, readError := os.Open(event.Name)
	defer csvFile.Close()
	if readError != nil {
		return output, readError
	}
	scanner := bufio.NewScanner(csvFile)
	output = append(output, scan_mutilate_stdout_rows(scanner, baseTime)...)

	return output, nil
}

func scan_mutilate_stdout_rows(scanner *bufio.Scanner, eventTime time.Time) []MutilateMetric {
	var output []MutilateMetric
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "read") {
			output = append(output, parse_mutilate_stdout_row_default(line, eventTime)...)
		} else if strings.Contains(line, "Swan latency for percentile") {
			output = append(output, parse_mutilate_stdout_row_swan(line, eventTime))
		}

	}

	return output
}

func parse_mutilate_stdout_row_default(line string, eventTime time.Time) []MutilateMetric {
	var output []MutilateMetric
	fields := strings.Fields(line)
	metricNames := get_default_metrics_names()
	for key, value := range fields {
		if key == 0 {
			continue
		}
		output = append(output,
			create_metrics_from_default_stdout(value, metricNames[key], eventTime))
	}

	return output
}

func get_default_metrics_names() []string {
	var names []string
	names = append(names, "", "avg", "std", "min", "percentile/5th", "percentile/10th",
		"percentile/90th", "percentile/95th", "percentile/99th")

	return names

}

func create_metrics_from_default_stdout(value string, name string, eventTime time.Time) MutilateMetric {
	floatValue, _ := strconv.ParseFloat(value, 64)
	metric := MutilateMetric{name, floatValue, eventTime}
	return metric
}

func parse_mutilate_stdout_row_swan(line string, eventTime time.Time) MutilateMetric {
	lineFields := strings.Split(line, ":")
	floatValue, _ := strconv.ParseFloat(strings.TrimSpace(lineFields[1]), 64)
	name := get_metric_name_from_swan_line_description(lineFields[0])
	output := MutilateMetric{name, floatValue, eventTime}

	return output
}

func get_metric_name_from_swan_line_description(description string) string {
	words := strings.Split(description, " ")
	percentileName := words[len(words)-1]
	var buffer bytes.Buffer
	buffer.WriteString("percentile/")
	buffer.WriteString(strings.Trim(percentileName, "0"))
	buffer.WriteString("th")

	return buffer.String()
}
