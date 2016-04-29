package experiment

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"os"
	"path"
	"strconv"
	"time"
)

// Experiment captures the internal data for the Experiment Driver.
type Experiment struct {
	session             session
	name                string
	workingDirectory    string
	measurements        []Measurement
	startingDirectory   string
	experimentDirectory string

	logFile *os.File
	logLvl  log.Level
	Log     *log.Logger
}

// NewExperiment creates a new Experiment instance,
// initialize experiment working directory and initialize logs.
// Caller have to provide slice of Measurement interfaces which are going to be executed.
func NewExperiment(name string, measurements []Measurement,
	directory string, logLevel log.Level) (*Experiment, error) {
	if len(measurements) == 0 {
		return nil, errors.New("invalid argument: no measurements provided")
	}

	// TODO(mpatelcz): Check if measurements names are unique!
	e := &Experiment{
		name:             name,
		session:          newSession(),
		workingDirectory: directory,
		measurements:     measurements,
		logLvl:           logLevel,
	}

	err := e.createExperimentDir()
	if err != nil {
		return nil, err
	}

	err = e.logInitialize()
	if err != nil {
		os.Chdir(e.startingDirectory)
		return nil, err
	}

	return e, nil
}

// Run runs experiment.
// It runs in a sequence defined measurements with given repetition count.
func (e *Experiment) Run() error {
	experimentStartingTime := time.Now()

	e.Log.Info("Starting Experiment ", e.name, " with uuid ", e.session.Name)
	var err error
	for _, measurement := range e.measurements {
		for i := 0; i < measurement.Repetitions(); i++ {
			measurementStartingTime := time.Now()

			e.Log.Info("Starting ", measurement.Name(), " repetition ", i)
			err = e.createMeasurementDir(measurement, i)
			if err != nil {
				return err
			}

			err := measurement.Run(e.Log)
			if err != nil {
				// When measurement return error we stop the whole experiment.
				e.Log.Error(measurement.Name(), " returned error ", err)
				return err
			}

			e.Log.Info("Ended ", measurement.Name(), " repetition ", i,
				" in ", time.Since(measurementStartingTime))
		}

		// Give a chance for measurement to finalize.
		// E.g to check statistical confidence of a result based on repetitions results.
		e.Log.Info("Finalizing ", measurement.Name(),
			" after ", measurement.Repetitions(), " repetitions")
		err := measurement.Finalize()
		if err != nil {
			// When measurement return error we stop the whole experiment.
			e.Log.Error(measurement.Name(), " returned error ", err,
				" while finalizing.")
			return err
		}
	}
	e.Log.Info("Ended experiment ", e.name, " with uuid ", e.session.Name,
		" in ", time.Since(experimentStartingTime))
	return err
}

// createExperimentDir creates unique directory for experiment logs and results.
func (e *Experiment) createExperimentDir() error {
	e.startingDirectory, _ = os.Getwd()

	if len(e.workingDirectory) > 0 {
		e.experimentDirectory = path.Join(e.workingDirectory, e.name, e.session.Name)
	} else {
		e.experimentDirectory = path.Join(e.startingDirectory, e.name, e.session.Name)
	}

	err := os.MkdirAll(e.experimentDirectory, 0777)
	if err != nil {
		return err
	}

	err = os.Chdir(e.experimentDirectory)
	return err
}

func (e *Experiment) Finalize() {
	e.logClose()

	// Exit experiment directory
	os.Chdir(e.startingDirectory)
}

func (e *Experiment) createMeasurementDir(measurement Measurement, iteration int) error {
	measurementDir := path.Join(e.experimentDirectory,
		measurement.Name(), strconv.FormatInt(int64(iteration), 10))

	err := os.MkdirAll(measurementDir, 0777)
	if err != nil {
		return err
	}

	err = os.Chdir(measurementDir)
	if err != nil {
		return err
	}

	return err
}

func (e *Experiment) logInitialize() error {
	var err error

	// Create master log file "master.log".
	e.logFile, err = os.OpenFile(e.experimentDirectory+"/master.log", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	// Create new logger.
	e.Log = &log.Logger{
		Out:       e.logFile,
		Formatter: new(log.TextFormatter),
		Hooks:     make(log.LevelHooks),
		Level:     e.logLvl,
	}

	return err
}

func (e *Experiment) logClose() {
	e.logFile.Close()
}
