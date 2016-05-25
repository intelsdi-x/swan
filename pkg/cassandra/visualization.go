package cassandra

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"strings"
)

func DrawTable(experimentID string) error {
	data := [][]string{}
	headers := []string{"namespace", "version", "host", "time", "boolval", "doubleval", "labels", "strval", "tags", "valtype"}

	cassandraConfig, err := CreateConfigWithSession("127.0.0.1", "snap")
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
