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

	qpsMap, err := getQPS(metricsList)
	if err != nil {
		return err
	}

	// Prepare headers for view.
	// TODO(Ala) Get number of load points from cassandra when they will be available there.
	// loadPointsNumber := getLoadPointNumber()
	loadPointsNumber := 10
	headers := createHeadersForSensitivityProfile(loadPointsNumber, qpsMap)

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
	scenarios := []string{}

	for _, metrics := range metricsList {
		scenarios = append(scenarios, createUniqueList("swan_aggressor_name", metrics.Tags(), scenarios)...)
	}

	// Create each row for scenario.
	for _, scenario := range scenarios {

		loadPointValues := map[int]string{}
		// Get all values for each aggressor from metrics.
		loadPointValues, err = getValuesForLoadPoints(metricsList, scenario)
		if err != nil {
			return nil, err
		}

		row := []string{}
		// Append values to row in correct order based on load point ID.
		for loadPoint := 1; loadPoint <= loadPointsNumber; loadPoint++ {
			row = append(row, loadPointValues[loadPoint])
		}

		// Append labels to rows, if aggressor is None change label to Baseline.
		// Append rows to table data, Baseline in a first row.
		if scenario == "None" {
			row = append([]string{"Baseline"}, row...)
			data = append([][]string{row}, data...)
		} else {
			row = append([]string{scenario}, row...)
			data = append(data, row)
		}
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

func createHeadersForSensitivityProfile(loadPointsNumber int, qpsMap map[int]string) (headers []string) {
	var qps string
	headers = append(headers, "Scenario/Load")
	// Calculate percentage for each load point - from 5% to 95 %.
	for loadPoint := 0; loadPoint < loadPointsNumber; loadPoint++ {
		if len(qpsMap[loadPoint+1]) > 0 {
			qps = qpsMap[loadPoint+1]
		} else {
			qps = ""
		}
		percentage := 5 + 90*loadPoint/(loadPointsNumber-1)
		headers = append(headers, (fmt.Sprintf("%d%% ", percentage) + "(" + qps + ")"))
	}
	return headers
}

func calculateAverage(valuesList []string) (result float64, err error) {
	if len(valuesList) == 0 {
		return result, errors.New("Empty list of values for given phase")
	}
	var sum float64
	for _, elem := range valuesList {
		value, err := strconv.ParseFloat(elem, 64)
		if err != nil {
			return result, err
		}
		sum += value
	}
	result = sum / float64(len(valuesList))
	return result, nil
}

// TODO(ala) For getting LoadPointID - replace with id gathered directly from metrics, when we add loadPointID there.
func getNumberForRegex(name string, regex string) (result int, err error) {
	re := regexp.MustCompile(regex)
	match := re.FindStringSubmatch(name)
	if len(match) == 0 {
		return result, fmt.Errorf("Could not retrieve number from string: %s", name)
	}
	number, err := strconv.Atoi(match[1])
	if err != nil {
		return result, err
	}
	return number, nil
}

func getQPS(metricsList []*cassandra.Metrics) (map[int]string, error) {
	qpsMap := make(map[int]string)
	var aggressor string

	// Get one scenario and find all qps values for it,
	// as qps values are the same for all scenarios,
	// they only vary for load points.
	if len(metricsList) > 0 {
		metric := metricsList[0]
		aggressor = metric.Tags()["swan_aggressor_name"]
	}
	for _, metrics := range metricsList {
		if metrics.Tags()["swan_aggressor_name"] == aggressor {
			// Load point ID is last digit in given phase ID, extract it and return.
			key, err := getNumberForRegex(metrics.Tags()["swan_phase"], `([0-9]+)$`)
			if err != nil {
				return nil, err
			}
			qpsMap[key] = metrics.Tags()["swan_loadpoint_qps"]

		}
	}
	return qpsMap, nil
}

func getValuesForLoadPoints(metricsList []*cassandra.Metrics, aggressor string) (map[int]string, error) {
	loadPointValues := make(map[int]string)
	allLoadPointValues := make(map[int][]string)

	for _, metrics := range metricsList {
		// In sensitivity profile we accept only double values.
		if metrics.Valtype() != "doubleval" {
			return nil, errors.New("Values for sensitivity profile should have double type.")
		}
		percentile, err := getNumberForRegex(metrics.Namespace(), `([0-9]+)th$`)
		if err != nil {
			return nil, err
		}
		// Currently we want to show 99th percentile, we can change it here in future.
		if percentile == 99 && metrics.Tags()["swan_aggressor_name"] == aggressor {

			// Find metric with phase ID and extract load point ID from it.
			// Add to map all values for key equals each load point ID.
			number, err := getNumberForRegex(metrics.Tags()["swan_phase"], `([0-9]+)$`)
			if err != nil {
				return nil, err
			}
			allLoadPointValues[number] = append(allLoadPointValues[number],
				fmt.Sprintf("%f", metrics.Doubleval()))

		}
	}

	// From all values for each load point calculate average value.
	for key, list := range allLoadPointValues {
		value, err := calculateAverage(list)
		if err != nil {
			return nil, err
		}
		loadPointValues[key] = fmt.Sprintf("%f", value)
	}

	return loadPointValues, nil
}
