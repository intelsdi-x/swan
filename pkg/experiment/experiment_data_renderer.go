package experiment

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	"github.com/intelsdi-x/swan/pkg/visualization"
)

// Draw prepares data for given experiment ID and address where Cassandra is running.
// It creates model of data in a form of table and asks view to draw it.
func Draw(experimentID string, cassandraAddr string) error {
	headers := []string{"namespace", "version", "host", "time", "value", "tags"}

	// Configure Cassandra connection.
	cassandraConfig, err := cassandra.CreateConfigWithSession(cassandraAddr, "snap")
	if err != nil {
		return err
	}

	// Get data from Cassandra.
	metricsList, err := cassandraConfig.GetValuesForGivenExperiment(experimentID)
	if err != nil {
		return err
	}

	// Prepare data for view.
	data := prepareData(metricsList)

	// View table.
	visualization.PrintExperimentMetadata(visualization.NewExperimentMetadata(experimentID))
	table := visualization.NewTable(headers, data)
	visualization.DrawTable(table)

	return nil
}

// List prints list of experimentIds on stdout from Cassandra running on given IP.
func List(cassandraIP string) (err error) {
	uniqueNames := []string{}
	tagsMapsList, err := cassandra.GetTags(cassandraIP)
	if err != nil {
		return err
	}
	for _, elem := range tagsMapsList {
		uniqueNames = append(uniqueNames, createUniqueList("swan_experiment", elem, uniqueNames)...)
	}
	visualization.PrintList(visualization.NewList(uniqueNames, "experiment ID:"))
	return nil
}

func mapToString(m map[string]string) (result string) {
	for key, value := range m {
		result += fmt.Sprintf("%s:%s\n", key, value)
	}
	return result
}

func getStringFromMetricValue(valtype string, metrics *cassandra.Metrics) (result string) {
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

func prepareData(metricsList []*cassandra.Metrics) (data [][]string) {
	for _, metrics := range metricsList {
		rowList := []string{}
		rowList = append(rowList, metrics.Namespace())
		rowList = append(rowList, fmt.Sprintf("%d", metrics.Version()))
		rowList = append(rowList, metrics.Host())
		rowList = append(rowList, metrics.Time().String())
		rowList = append(rowList, getStringFromMetricValue(metrics.Valtype(), metrics))
		rowList = append(rowList, mapToString(metrics.Tags()))
		data = append(data, rowList)
	}
	return data
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

func createUniqueList(key string, mapWithValues map[string]string, uniqueNames []string) (returnedNames []string) {
	// Add new value from map to uniqueNames if it does not exist in given uniqueNames for given key.
	for k, value := range mapWithValues {
		if k == key && !isValueInSlice(value, uniqueNames) {
			returnedNames = append(returnedNames, value)
		}
	}
	return returnedNames
}
