package experiment

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
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
	phases              []Phase
	startingDirectory   string
	experimentDirectory string

	logFile  *os.File
	logLevel log.Level
}

// NewExperiment creates a new Experiment instance,
// initialize experiment working directory and initialize logs.
// Caller have to provide slice of Phase interfaces which are going to be executed.
func NewExperiment(name string, phases []Phase,
	directory string, logLevel log.Level) (*Experiment, error) {
	if len(phases) == 0 {
		return nil, errors.New("invalid argument: no phases provided")
	}

	// TODO(mpatelcz): Check if phases names are unique!
	e := &Experiment{
		name:             name,
		session:          newSession(),
		workingDirectory: directory,
		phases:           phases,
		logLevel:         logLevel,
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
// It runs in a sequence defined phases with given repetition count.
func (e *Experiment) Run() error {
	experimentStartingTime := time.Now()

	log.Info("Starting Experiment ", e.name, " with uuid ", e.session.Name)
	var err error
	for _, phase := range e.phases {
		for i := 0; i < phase.Repetitions(); i++ {
			// TODO: Trigger snap session here. Fetch workflow & config from Phase.
			// (proposition) and add session name & experiment id tags.

			phaseStartingTime := time.Now()

			log.Info("Starting ", phase.Name(), " repetition ", i)
			err = e.createPhaseDir(phase, i)
			if err != nil {
				return err
			}

			err := phase.Run()
			if err != nil {
				// When phase return error we stop the whole experiment.
				log.Error(phase.Name(), " returned error ", err)
				return err
			}

			log.Info("Ended ", phase.Name(), " repetition ", i,
				" in ", time.Since(phaseStartingTime))
		}

		// Give a chance for phase to finalize.
		// E.g to check statistical confidence of a result based on repetitions results.
		log.Info("Finalizing ", phase.Name(),
			" after ", phase.Repetitions(), " repetitions")
		err := phase.Finalize()
		if err != nil {
			// When phase return error we stop the whole experiment.
			log.Error(phase.Name(), " returned error ", err,
				" while finalizing.")
			return err
		}
	}
	log.Info("Ended experiment ", e.name, " with uuid ", e.session.Name,
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

// Finalize closes log file and returns to the previous working directory.
func (e *Experiment) Finalize() {
	e.logClose()

	// Exit experiment directory
	os.Chdir(e.startingDirectory)
}

func (e *Experiment) createPhaseDir(phase Phase, iteration int) error {
	phaseDir := path.Join(e.experimentDirectory,
		phase.Name(), strconv.FormatInt(int64(iteration), 10))

	err := os.MkdirAll(phaseDir, 0777)
	if err != nil {
		return err
	}

	err = os.Chdir(phaseDir)
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

	// Setup logging set to both output and logFile.
	log.SetLevel(e.logLevel)
	log.SetFormatter(new(log.TextFormatter))
	log.SetOutput(io.MultiWriter(e.logFile, os.Stdout))

	return err
}

func (e *Experiment) logClose() {
	e.logFile.Close()
}
