package cassandra

import (
	"github.com/vektra/errors"
	"time"
)

// GetValuesForGivenExperiment returns list of Metrics structs based on experiment ID.
func (cassandraConfig *Connection) GetValuesForGivenExperiment(experimentID string) (metrics []*Metrics, err error) {
	metricsList := []*Metrics{}
	// We look for following values for each Metric.
	var doubleval float64
	var namespace, host, strval, valtype string
	var version int
	var time time.Time
	var boolval bool
	tags := make(map[string]string)

	// Get current cassandra session and select values.
	session := cassandraConfig.CassandraSession()
	iter := session.Query(`SELECT ns, ver, host, time, boolval, doubleval, strval, tags, valtype FROM snap.metrics
	WHERE tags CONTAINS '` + experimentID + `'ALLOW FILTERING`).Iter()

	if len(iter.Columns()) > 0 {
		// Iterate through all gathered row, create Metrics struct for each row and add it to a list.
		for iter.Scan(&namespace, &version, &host, &time, &boolval, &doubleval, &strval, &tags, &valtype) {
			metric := NewMetrics(namespace, version, host, time, boolval, doubleval, strval, tags, valtype)
			metricsList = append(metricsList, metric)
		}
		return metricsList, nil
	}
	return nil, errors.New("Bad columns names")
}
