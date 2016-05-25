package cassandra

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"strings"
)

// DrawTable draws table for given experiment Id.
func DrawTable(experimentID string, host string) error {
	data := [][]string{}
	headers := []string{"namespace", "version", "host", "time", "boolval", "doubleval", "labels", "strval", "tags", "valtype"}

	cassandraConfig, err := CreateConfigWithSession(host, "snap")
	if err != nil {
		return err
	}
	metricsList := cassandraConfig.GetValuesForGivenExperiment(experimentID)
	fmt.Println("\n")
	fmt.Println("Experiment id: " + experimentID)
	for _, metrics := range metricsList {
		rowList := []string{}
		rowList = append(rowList, metrics.Namespace())
		rowList = append(rowList, fmt.Sprintf("%d", metrics.Version()))
		rowList = append(rowList, metrics.Host())
		rowList = append(rowList, metrics.Time().String())
		rowList = append(rowList, fmt.Sprintf("%t", metrics.Boolval()))
		rowList = append(rowList, fmt.Sprintf("%f", metrics.Doubleval()))
		rowList = append(rowList, strings.Join(metrics.Labels()[:], ","))
		rowList = append(rowList, metrics.Strval())
		rowList = append(rowList, fmt.Sprintf("%+v", metrics.Tags()))
		rowList = append(rowList, metrics.Valtype())
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
	iter := cassandraConfig.session.Query(`SELECT tags FROM snap.metrics`).Iter()
	for iter.Scan(&tagsMap) {
		tagsMapsList = append(tagsMapsList, tagsMap)
	}

	return tagsMapsList, err
}

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
