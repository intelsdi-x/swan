package cassandra

import (
	"github.com/gocql/gocql"
)

func GetValuesAndTagsForGivenExperiment(session *gocql.Session, experimentName string) (valuesList []float64,
	tagsList []map[string]string) {
	var value float64
	tagsMap := make(map[string]string)
	iter := session.Query(`SELECT doubleval, tags FROM snap.metrics WHERE tags CONTAINS '` + experimentName +
		`'ALLOW FILTERING`).Iter()
	for iter.Scan(&value, &tagsMap) {
		valuesList = append(valuesList, value)
		tagsList = append(tagsList, tagsMap)
	}
	defer session.Close()
	return valuesList, tagsList
}
