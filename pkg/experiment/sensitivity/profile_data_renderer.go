package sensitivity

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	"github.com/intelsdi-x/swan/pkg/visualization"
	"regexp"
	"strconv"
)

// Draw prepares data for sensitivity table with values for each aggressor and load point for given experiment ID and
// Cassandra running on given IP.
// It creates model of data in a form of table and asks view to draw it.
func Draw(experimentID string, cassandraAddr string) error {
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

	// Prepare headers for view.
	// TODO(Ala) Get number of load points from cassandra when they will be available there.
	// loadPointsNumber := getLoadPointNumber()
	loadPointsNumber := 10
	headers := createHeadersForSensitivityProfile(loadPointsNumber)

	// Prepare data for view.
	data, err := prepareData(metricsList, loadPointsNumber)
	if err != nil {
		return err
	}

	// View table.
	visualization.PrintExperimentMetadata(visualization.NewExperimentMetadata(experimentID))
	table := visualization.NewTable(headers, data)
	visualization.DrawTable(table)
	return nil
}

func prepareData(metricsList []*cassandra.Metrics, loadPointsNumber int) (data [][]string, err error) {
	// List of unique aggressors names for given experiment ID.
	aggressors := []string{}

	for _, metrics := range metricsList {
		aggressors = append(aggressors, createUniqueList("swan_aggressor_name", metrics.Tags(), aggressors)...)
	}

	// Create each row for aggressor.
	for _, aggressor := range aggressors {
		loadPointValues := map[int]string{}
		// Get all values for each aggressor from metrics.
		loadPointValues, err = getValuesForLoadPoints(metricsList, aggressor)
		if err != nil {
			return nil, err
		}
		rowList := []string{}
		rowList = append(rowList, aggressor)
		// Append values to row in correct order based on load point ID.
		for loadPoint := 1; loadPoint < loadPointsNumber; loadPoint++ {
			rowList = append(rowList, loadPointValues[loadPoint])
		}
		data = append(data, rowList)
	}
	return data, err
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

// TODO(ala) Replace with id gathered directly from metrics, when we add loadPointID there.
func getLoadPointNumber(phase string) (*int, error) {
	// Load point ID is last digit in given phase ID, extract it and return.
	re := regexp.MustCompile(`([0-9]+)$`)
	match := re.FindStringSubmatch(phase)
	if len(match) == 0 {
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
						fmt.Sprintf("%f", metrics.Doubleval()))
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
