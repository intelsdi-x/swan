package experiment

import (
	"errors"
	"log"
	"os"
	"path"
	"strconv"
)

// Experiment captures the internal data for the Experiment Driver.
type Experiment struct {
	session             session
	name                string
	workingDirectory    string
	phases              []Phase
	startingDirectory   string
	experimentDirectory string
	logFile             *os.File
	logger              *log.Logger
}

// NewExperiment creates a new Experiment instance,
// initialize experiment working directory and initialize logs.
// Caller have to provide slice of Phase interfaces which are going to be executed.
func NewExperiment(name string, phases []Phase, directory string) (*Experiment, error) {
	if len(phases) == 0 {
		return nil, errors.New("invalid argument: no phases provided")
	}

	// TODO(mpatelcz): Check if phase names are unique!

	e := &Experiment{
		name:             name,
		session:          newSession(),
		workingDirectory: directory,
		phases:           phases,
	}

	err := e.createExperimentDir()
	if err != nil {
		return nil, err
	}

	err = e.logInitialize()
	if err != nil {
		return nil, err
	}

	return e, nil
}

// Run runs experiment.
// It runs in a sequence defined phases with given repetition count.
func (e *Experiment) Run() error {
	var err error

	// Always perform cleanup on exit.
	defer e.finalize()

	e.logger.Printf("starting experiment '%s' with uuid '%s'\n", e.name, e.session.Name)

	for _, phase := range e.phases {
		for i := 0; i < phase.Repetitions(); i++ {
			e.logger.Printf("starting phase '%s' iteration %d\n", phase.Name(), i)

			err = e.createPhaseDir(phase, i)
			if err != nil {
				return err
			}

			err := phase.Run()
			if err != nil {
				e.logger.Printf("phase returned error '%v'\n", err)
				return err
			}

			e.logger.Printf("ended phase '%s' iteration %d\n", phase.Name(), i)
		}
	}

	e.logger.Printf("ended experiment '%s' with uuid '%s'\n", e.name, e.session.Name)

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

func (e *Experiment) finalize() {
	e.logClose()

	// Exit experiment directory
	os.Chdir(e.startingDirectory)
}

func (e *Experiment) createPhaseDir(phase Phase, iteration int) error {
	phaseDir := path.Join(e.experimentDirectory, phase.Name(), strconv.FormatInt(int64(iteration), 10))

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

	// create master log file "master.log"
	e.logFile, err = os.Create(e.experimentDirectory + "/master.log")
	if err != nil {
		os.Chdir(e.startingDirectory)
		return err
	}

	e.logger = log.New(e.logFile, "", log.LstdFlags)
	if e.logger == nil {
		os.Chdir(e.startingDirectory)
		return errors.New("failed to create master log file")
	}

	return err
}

func (e *Experiment) logClose() {
	e.logger.SetOutput(os.Stdout)
	e.logFile.Close()
}
