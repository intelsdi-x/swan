package testhelpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	
	"github.com/pkg/errors"
)

// FilePublisherMetric is used for decoding content of file publisher.
type FilePublisherMetric struct {
	// Namespace contains snap metric namespace.
	Namespace string `json:"namespace"`
	// Data contains snap metric value.
	Data interface{} `json:"data"`
	// Tags contains map of tags from metric.
	Tags map[string]string `json:"tags"`
}

// GetMetric returns FilePublisherMetric structure with requested metric's namespace.
func GetMetric(namespace string, metrics []FilePublisherMetric) (*FilePublisherMetric, error) {
	for _, metric := range metrics {
		if metric.Namespace == namespace {
			return &metric, nil
		}
	}
	return nil, fmt.Errorf("cannot find specified metric")
}

// GetOneMeasurementFromFile gets one random measurement from file.
func GetOneMeasurementFromFile(fileLocation string) ([]FilePublisherMetric, error) {
	var oneMeasurement []FilePublisherMetric

	content, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return nil, err
	}
	measurements := strings.Split(string(content), "\n")

	if err := json.Unmarshal([]byte(measurements[0]), &oneMeasurement); err != nil {
		return nil, errors.Wrapf(err, "cannot parse output file measurments: %v", measurements)
	}

	return oneMeasurement, nil
}
