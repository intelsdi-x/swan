package cassandra

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

func mapToString(m map[string]string) (result string) {
	for key, value := range m {
		result += fmt.Sprintf("%s:%s\n", key, value)
	}
	return result
}

func getMetricForValtype(valtype string, metrics *Metrics) (result string) {
	switch valtype {
	case "boolval":
		result = fmt.Sprintf("%t", metrics.Boolval())
	case "strval":
		result = metrics.Strval()
	case "doubleval":
		result = fmt.Sprintf("%f", metrics.Doubleval())
	}
	return result
}

// DrawTable draws table for given experiment ID.
func DrawTable(experimentID string, host string) error {
	data := [][]string{}
	headers := []string{"namespace", "version", "host", "time", "value", "tags"}

	cassandraConfig, err := CreateConfigWithSession(host, "snap")
	if err != nil {
		return err
	}

	metricsList, err := cassandraConfig.GetValuesForGivenExperiment(experimentID)
	if err != nil {
		return err
	}

	fmt.Println("\n")
	fmt.Println("Experiment id: " + experimentID)
	for _, metrics := range metricsList {
		// TODO(akwasnie) filter columns to show only some of them.
		rowList := []string{}
		rowList = append(rowList, metrics.Namespace())
		rowList = append(rowList, fmt.Sprintf("%d", metrics.Version()))
		rowList = append(rowList, metrics.Host())
		rowList = append(rowList, metrics.Time().String())
		rowList = append(rowList, getMetricForValtype(metrics.Valtype(), metrics))
		rowList = append(rowList, mapToString(metrics.Tags()))
		data = append(data, rowList)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	for _, v := range data {
		table.Append(v)
	}
	table.Render()

	return nil
}

func getTags(host string) (tagsMapsList []map[string]string, err error) {
	var tagsMap map[string]string
	cassandraConfig, err := CreateConfigWithSession(host, "snap")
	if err != nil {
		return nil, err
	}

	iter := cassandraConfig.session.Query(`SELECT tags FROM snap.metrics`).Iter()

	for iter.Scan(&tagsMap) {
		tagsMapsList = append(tagsMapsList, tagsMap)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return tagsMapsList, nil
}

// isValueInSlice is used to check whether given string already exists in given slice.
func isValueInSlice(value string, slice []string) bool {
	for _, elem := range slice {
		if elem == value {
			return true
		}
	}
	return false
}

// DrawList returns list of experimentIds.
func DrawList(host string) (err error) {
	var uniqueNames []string
	tagsMapsList, err := getTags(host)
	if err != nil {
		return err
	}

	for _, elem := range tagsMapsList {
		for k, value := range elem {
			if k == "swan_experiment" && !isValueInSlice(value, uniqueNames) {
				uniqueNames = append(uniqueNames, value)
				fmt.Println(value)
			}
		}
	}
	return nil
}
