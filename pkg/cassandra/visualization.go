package cassandra

import (
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"regexp"
	"strconv"
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

func createUniqueList(key string, elem map[string]string, uniqueNames []string) (returnedNames []string) {
	for k, value := range elem {
		if k == key && !isValueInSlice(value, uniqueNames) {
			returnedNames = append(returnedNames, value)
		}
	}
	return returnedNames
}

// DrawList returns list of experimentIds.
func DrawList(host string) (err error) {
	uniqueNames := []string{}
	tagsMapsList, err := getTags(host)
	if err != nil {
		return err
	}
	for _, elem := range tagsMapsList {
		uniqueNames = append(uniqueNames, createUniqueList("swan_experiment", elem, uniqueNames)...)
	}
	for _, value := range uniqueNames {
		fmt.Println(value)
	}
	return nil
}

func createHeadersForSensitivityProfile() (headers []string) {
	headers = append(headers, "Aggressor name")
	// TODO(Ala) Get number of load points from cassandra when they will be available there.
	// loadPointsNumber := getLoadPointNumber()
	loadPointsNumber := 10
	for loadPoint := 0; loadPoint < loadPointsNumber; loadPoint++ {
		percentage := 5 + 90*loadPoint/(loadPointsNumber-1)
		headers = append(headers, fmt.Sprintf("load point %d%%", percentage))
	}
	return headers
}

func calculateAverage(valuesList []string) (*float64, error) {
	if len(valuesList) == 0 {
		return nil, errors.New("Empty list of values for given phase")
	}
	var sum float64
	for _, elem := range valuesList {
		value, err := strconv.ParseFloat(elem, 64)
		if err != nil {
			return nil, err
		}
		sum += value
	}
	result := sum / float64(len(valuesList))
	return &result, nil
}

func getValuesForLoadPoints(metricsList []*Metrics, aggressor string) (map[int]string, error) {
	loadPointValues := make(map[int]string)
	allLoadPointValues := make(map[int][]string)

	for _, metrics := range metricsList {
		if metrics.Tags()["swan_aggressor_name"] == aggressor {
			for key, value := range metrics.Tags() {
				if key == "swan_phase" {
					re := regexp.MustCompile(`([0-9]+)$`)
					match := re.FindStringSubmatch(value)
					if len(match[1]) == 0 {
						errorMsg := fmt.Sprintf(
							"Could not retrieve load point number from phase: %s", value)
						return nil, errors.New(errorMsg)
					}
					number, err := strconv.Atoi(match[1])
					if err != nil {
						return nil, err
					}
					allLoadPointValues[number] = append(allLoadPointValues[number],
						getMetricForValtype(metrics.Valtype(), metrics))
				}
			}
		}
	}

	for key, list := range allLoadPointValues {
		value, err := calculateAverage(list)
		if err != nil {
			return nil, err
		}
		loadPointValues[key] = fmt.Sprintf("%f", *value)
	}

	return loadPointValues, nil
}

// DrawSensitivityProfile draws table with values for each aggressor and load point for given experiment ID.
func DrawSensitivityProfile(experimentID string, host string) error {
	data := [][]string{}
	aggressors := []string{}
	headers := createHeadersForSensitivityProfile()

	// TODO(Ala) Get number of load points from cassandra when they will be available there.
	// loadPointsNumber := getLoadPointNumber()
	loadPointsNumber := 10

	cassandraConfig, err := CreateConfigWithSession(host, "snap")
	if err != nil {
		return err
	}
	metricsList, err := cassandraConfig.GetValuesForGivenExperiment(experimentID)
	if err != nil {
		return err
	}

	for _, metrics := range metricsList {
		aggressors = append(aggressors, createUniqueList("swan_aggressor_name", metrics.Tags(), aggressors)...)
	}

	for _, aggressor := range aggressors {
		loadPointValues := map[int]string{}
		loadPointValues, err = getValuesForLoadPoints(metricsList, aggressor)
		if err != nil {
			return err
		}
		rowList := []string{}
		rowList = append(rowList, aggressor)
		for loadPoint := 1; loadPoint < loadPointsNumber; loadPoint++ {
			rowList = append(rowList, loadPointValues[loadPoint])
		}
		data = append(data, rowList)
	}

	fmt.Println("\n")
	fmt.Println("Experiment id: " + experimentID)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
	return nil
}
