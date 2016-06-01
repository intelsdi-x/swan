package sensitivity

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	"github.com/intelsdi-x/swan/pkg/visualization"
	"regexp"
	"strconv"
)

func mapToString(m map[string]string) (result string) {
	for key, value := range m {
		result += fmt.Sprintf("%s:%s\n", key, value)
	}
	return result
}

func getMetricForValtype(valtype string, metrics *cassandra.Metrics) (result string) {
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

// ExperimentTable draws table for given experiment ID.
func ExperimentTable(experimentID string, host string) error {
	data := [][]string{}
	headers := []string{"namespace", "version", "host", "time", "value", "tags"}

	cassandraConfig, err := cassandra.CreateConfigWithSession(host, "snap")
	if err != nil {
		return err
	}

	metricsList, err := cassandraConfig.GetValuesForGivenExperiment(experimentID)
	if err != nil {
		return err
	}

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
	visualization.PrintExperimentMetadata(visualization.NewExperimentMetadata(experimentID))
	table := visualization.NewTable(headers, data)
	visualization.DrawTable(table)
	return nil
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
	// Add new value from map to uniqueNames if it does not exist in given uniqueNames.
	for k, value := range elem {
		if k == key && !isValueInSlice(value, uniqueNames) {
			returnedNames = append(returnedNames, value)
		}
	}
	return returnedNames
}

// List prints list of experimentIds on stdout.
func List(host string) (err error) {
	uniqueNames := []string{}
	tagsMapsList, err := cassandra.GetTags(host)
	if err != nil {
		return err
	}
	for _, elem := range tagsMapsList {
		uniqueNames = append(uniqueNames, createUniqueList("swan_experiment", elem, uniqueNames)...)
	}
	visualization.PrintList(uniqueNames)
	return nil
}

func createHeadersForSensitivityProfile(loadPointsNumber int) (headers []string) {
	headers = append(headers, "Scenario/Load")
	// Calculate percentage for each load point - from 5% to 95 %.
	for loadPoint := 0; loadPoint < loadPointsNumber; loadPoint++ {
		percentage := 5 + 90*loadPoint/(loadPointsNumber-1)
		headers = append(headers, fmt.Sprintf("%d%%", percentage))
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

func getLoadPointNumber(phase string) (*int, error) {
	// Load point ID is last digit in given phase ID, extract it and return.
	re := regexp.MustCompile(`([0-9]+)$`)
	match := re.FindStringSubmatch(phase)
	if len(match[1]) == 0 {
		errorMsg := fmt.Sprintf(
			"Could not retrieve load point number from phase: %s", phase)
		return nil, errors.New(errorMsg)
	}
	number, err := strconv.Atoi(match[1])
	if err != nil {
		return nil, err
	}
	return &number, nil
}

func getValuesForLoadPoints(metricsList []*cassandra.Metrics, aggressor string) (map[int]string, error) {
	loadPointValues := make(map[int]string)
	allLoadPointValues := make(map[int][]string)

	for _, metrics := range metricsList {
		// In sensitivity profile we accept only double values.
		if metrics.Valtype() != "doubleval" {
			return nil, errors.New("Values for sensitivity profile should have double type.")
		}
		if metrics.Tags()["swan_aggressor_name"] == aggressor {
			// Find metric with phase ID and extract load point ID from it.
			// Add to map all values for key equals each load point ID.
			for key, value := range metrics.Tags() {
				if key == "swan_phase" {
					number, err := getLoadPointNumber(value)
					if err != nil {
						return nil, err
					}
					allLoadPointValues[*number] = append(allLoadPointValues[*number],
						getMetricForValtype(metrics.Valtype(), metrics))
				}
			}
		}
	}

	// From all values for each load point calculate average value.
	for key, list := range allLoadPointValues {
		value, err := calculateAverage(list)
		if err != nil {
			return nil, err
		}
		loadPointValues[key] = fmt.Sprintf("%f", *value)
	}

	return loadPointValues, nil
}

// Profile draws sensitivity table with values for each aggressor and load point for given experiment ID.
func Profile(experimentID string, host string) error {
	// All table data.
	data := [][]string{}
	// List of unique aggressors names for given experiment ID.
	aggressors := []string{}

	// TODO(Ala) Get number of load points from cassandra when they will be available there.
	// loadPointsNumber := getLoadPointNumber()
	loadPointsNumber := 10
	headers := createHeadersForSensitivityProfile(loadPointsNumber)

	cassandraConfig, err := cassandra.CreateConfigWithSession(host, "snap")
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

	// Create each row for aggressor.
	for _, aggressor := range aggressors {
		loadPointValues := map[int]string{}
		// Get all values for each aggressor from metrics.
		loadPointValues, err = getValuesForLoadPoints(metricsList, aggressor)
		if err != nil {
			return err
		}
		rowList := []string{}
		rowList = append(rowList, aggressor)
		// Append values to row in correct order based on load point ID.
		for loadPoint := 1; loadPoint < loadPointsNumber; loadPoint++ {
			rowList = append(rowList, loadPointValues[loadPoint])
		}
		data = append(data, rowList)
	}
	visualization.PrintExperimentMetadata(visualization.NewExperimentMetadata(experimentID))
	table := visualization.NewTable(headers, data)
	visualization.DrawTable(table)
	return nil
}
