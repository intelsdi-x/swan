package sensitivity

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/intelsdi-x/swan/pkg/cassandra"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/visualization"
	"github.com/montanaflynn/stats"
	"github.com/pkg/errors"
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
	fmt.Println(visualization.NewExperimentMetadata(experimentID).String())
	table := visualization.NewTable(headers, data)
	table.Draw()
	return nil
}

func prepareData(metricsList []*cassandra.Metrics, loadPointsNumber int) ([][]string, error) {
	data := [][]string{}
	// List of unique aggressors names for given experiment ID.
	scenarios := []string{}

	for _, metrics := range metricsList {
		scenarios = append(scenarios, createUniqueList(phase.AggressorNameKey, metrics.Tags(), scenarios)...)
	}

	// Create each row for scenario.
	for _, scenario := range scenarios {
		loadPointValues := map[int]string{}
		// Get all values for each aggressor from metrics.
		loadPointValues, err := getValuesForLoadPoints(metricsList, scenario)
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
	return data, nil
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

func createUniqueList(key string, elem map[string]string, uniqueNames []string) []string {
	returnedNames := []string{}
	// Add new value from map to uniqueNames if it does not exist in given uniqueNames.
	for k, value := range elem {
		if k == key && !isValueInSlice(value, uniqueNames) {
			returnedNames = append(returnedNames, value)
		}
	}
	return returnedNames
}

func createHeadersForSensitivityProfile(loadPointsNumber int, qpsMap map[int]string) []string {
	headers := []string{}
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

// TODO(ala) For getting LoadPointID - replace with id gathered directly from metrics, when we add loadPointID there.
func getNumberForRegex(name string, regex string) (int, error) {
	var result int
	re := regexp.MustCompile(regex)
	match := re.FindStringSubmatch(name)
	if len(match) == 0 {
		return result, errors.Errorf("could not retrieve number from string %q", name)
	}
	number, err := strconv.Atoi(match[1])
	if err != nil {
		return result, errors.Wrap(err, "error while creating sensitivity profile")
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
		aggressor = metric.Tags()[phase.AggressorNameKey]
	}
	for _, metrics := range metricsList {
		if metrics.Tags()[phase.AggressorNameKey] == aggressor {
			// Load point ID is last digit in given phase ID, extract it and return.
			key, err := getNumberForRegex(metrics.Tags()[phase.PhaseKey], `([0-9]+)$`)
			if err != nil {
				return nil, err
			}
			qpsMap[key] = metrics.Tags()[phase.LoadPointQPSKey]

		}
	}
	return qpsMap, nil
}

func getValuesForLoadPoints(metricsList []*cassandra.Metrics, aggressor string) (map[int]string, error) {
	loadPointValues := make(map[int]string)
	allLoadPointValues := make(map[int][]float64)

	for _, metrics := range metricsList {
		// In sensitivity profile we accept only double values.
		if metrics.Valtype() != "doubleval" {
			return nil, errors.New("values for sensitivity profile should have type 'double'")
		}
		splittedNamespace := strings.Split(metrics.Namespace(), "/")
		if len(splittedNamespace) == 0 {
			return nil, errors.Errorf("bad namespace format %q", metrics.Namespace())
		}
		// Currently we want to show 99th percentile, we can change it here in future.
		if splittedNamespace[len(splittedNamespace)-1] == "99th" &&
			metrics.Tags()[phase.AggressorNameKey] == aggressor {

			// Find metric with phase ID and extract load point ID from it.
			// Add to map all values for key equals each load point ID.
			number, err := getNumberForRegex(metrics.Tags()[phase.PhaseKey], `([0-9]+)$`)
			if err != nil {
				return nil, err
			}
			allLoadPointValues[number] = append(allLoadPointValues[number], metrics.Doubleval())
		}
	}

	// From all values for each load point calculate average value.
	for key, list := range allLoadPointValues {
		stdev, err := stats.StandardDeviation(list)
		if err != nil {
			return nil, errors.Wrap(err, "standard deviation computation failed")
		}

		mean, err := stats.Mean(list)
		if err != nil {
			return nil, errors.Wrap(err, "mean computation failed")
		}

		loadPointValues[key] = fmt.Sprintf("%f (+/- %f)", mean, stdev)
	}

	return loadPointValues, nil
}
