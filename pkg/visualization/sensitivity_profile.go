package visualization

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/olekukonko/tablewriter"
	"github.com/vektra/errors"
	"os"
)

func configureCluster(ip string, keyspace string) *gocql.ClusterConfig {
	cluster := gocql.NewCluster(ip)
	cluster.Keyspace = keyspace
	cluster.ProtoVersion = 4
	cluster.Consistency = gocql.All
	return cluster
}

func getValuesForSpecificExperimentAndPhase(cluster *gocql.ClusterConfig, experimentName string, phaseName string) (
	valuesList []float64, err error) {
	var value float64
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	iter := session.Query(`SELECT doubleval FROM snap.metrics WHERE tags CONTAINS '` + experimentName +
		`' AND tags CONTAINS '` + phaseName + `'ALLOW FILTERING`).Iter()
	for iter.Scan(&value) {
		valuesList = append(valuesList, value)
	}
	return valuesList, err
}

func getTags(cluster *gocql.ClusterConfig) (tagsMapsList []map[string]string, err error) {
	var tagsMap map[string]string
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	iter := session.Query(`SELECT tags FROM snap.metrics`).Iter()
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

func extractUniqueTagsNames(tagsMapsList []map[string]string, tagName string) (uniqueNames []string) {
	for _, elem := range tagsMapsList {
		for k, value := range elem {
			if k == tagName && !isValueInSlice(value, uniqueNames) {
				uniqueNames = append(uniqueNames, value)
			}
		}
	}
	return uniqueNames
}

func calculateAverageForMeasurement(valuesList []float64) (result float64) {
	for _, value := range valuesList {
		result = result + value
	}
	if len(valuesList) > 0 {
		return result / float64(len(valuesList))
	}
	panic(errors.New("Array length has to be longer than 0"))
}

func drawTable(cluster *gocql.ClusterConfig) error {
	data := [][]string{}
	headers := []string{}
	tagsMapsList, err := getTags(cluster)
	if err != nil {
		return err
	}
	experimentNames := extractUniqueTagsNames(tagsMapsList, "swan_experiment")
	phasesNames := extractUniqueTagsNames(tagsMapsList, "swan_phase")
	for _, experimentName := range experimentNames {
		rowList := []string{}
		for _, phaseName := range phasesNames {
			headers = append(headers, phaseName)
			valuesList, err := getValuesForSpecificExperimentAndPhase(cluster, experimentName, phaseName)
			if err != nil {
				return err
			}
			rowList = append(rowList, fmt.Sprintf("%f", calculateAverageForMeasurement(valuesList)))
		}
		data = append(data, rowList)
		fmt.Println("\n")
		fmt.Println("Experiment name: " + experimentName)
		table := tablewriter.NewWriter(os.Stdout)

		table.SetHeader(headers)
		for _, v := range data {
			table.Append(v)
		}
		table.Render()
	}

	return nil
}
