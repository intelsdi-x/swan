package experiment

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	experimentPhase "github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/nu7hatch/gouuid"
	"io"
	"os"
	"path"
	"strconv"
	"time"
)

// Experiment captures the internal data for the Experiment Driver.
type Experiment struct {
	customName          string
	uuidName            string
	workingDirectory    string
	phases              []experimentPhase.Phase
	startingDirectory   string
	experimentDirectory string

	logFile  *os.File
	logLevel log.Level
}

// NewExperiment creates a new Experiment instance,
// initialize experiment working directory and initialize logs.
// Caller have to provide slice of Phase interfaces which are going to be executed.
func NewExperiment(customName string, phases []experimentPhase.Phase,
	directory string, logLevel log.Level) (*Experiment, error) {
	if len(phases) == 0 {
		return nil, errors.New("invalid argument: no phases provided")
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	// TODO(mpatelcz): Check if phases names are unique!
	e := &Experiment{
		customName:       customName,
		uuidName:         time.Now().Format("2006-01-02T15h04m05s_") + uuid.String(),
		workingDirectory: directory,
		phases:           phases,
		logLevel:         logLevel,
	}

	err = e.createExperimentDir()
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
func (e *Experiment) Run() (err error) {
	experimentStartingTime := time.Now()

	log.Info("Starting Experiment ", e.customName, " with uuid ", e.uuidName)
	for _, phase := range e.phases {
		for i := uint(0); i < phase.Repetitions(); i++ {
			// Create phase session.
			session := experimentPhase.Session{
				Name:         phase.Name(),
				ExperimentID: e.uuidName, // TODO: Decide if we want to have `customName` here as well.
				RepetitionID: i,
			}

			// Start timer.
			phaseStartingTime := time.Now()
			log.Info("Starting ", session.Name, " repetition ", session.RepetitionID)

			// Create and step into unique phase dir.
			err = e.createPhaseDir(session)
			if err != nil {
				return err
			}

			// Start phase.
			err := phase.Run(session)
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

	log.Info("Ended experiment ", e.customName, " with uuid ", e.uuidName,
		" in ", time.Since(experimentStartingTime))
	return err
}

// createExperimentDir creates unique directory for experiment logs and results.
func (e *Experiment) createExperimentDir() error {
	if len(e.workingDirectory) > 0 {
		e.startingDirectory = e.workingDirectory
	} else {
		e.startingDirectory, _ = os.Getwd()
	}

	e.experimentDirectory = path.Join(e.startingDirectory, e.customName, e.uuidName)

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

func (e *Experiment) createPhaseDir(session experimentPhase.Session) error {
	phaseDir := path.Join(e.experimentDirectory,
		session.Name, strconv.FormatInt(int64(session.RepetitionID), 10))

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
