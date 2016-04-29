package experiment

import "github.com/Sirupsen/logrus"

// Measurement defines interface which shall be provided by user for the
// Experiment Driver.
type Measurement interface {
	// Name returns measurement name.
	Name() string
	// Repetitions returns desired number of measurement repetitions.
	Repetitions() int
	// Run runs a measurement.
	Run(*logrus.Logger) error
	// Finalize is executed after all repetitions of given measurement.
	Finalize() error
}
