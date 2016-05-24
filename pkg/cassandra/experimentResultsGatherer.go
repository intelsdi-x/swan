package cassandra

// GetValuesAndTagsForGivenExperiment returns list of values and list of maps for tags based on experiment name.
func (cassandraConfig *Connection) GetValuesAndTagsForGivenExperiment(experimentName string) (valuesList []float64,
	tagsList []map[string]string) {
	var value float64
	tagsMap := make(map[string]string)
	session := cassandraConfig.CassandraSession()
	iter := session.Query(`SELECT doubleval, tags FROM snap.metrics WHERE tags CONTAINS '` + experimentName +
		`'ALLOW FILTERING`).Iter()
	for iter.Scan(&value, &tagsMap) {
		valuesList = append(valuesList, value)
		tagsList = append(tagsList, tagsMap)
	}
	return valuesList, tagsList
}
