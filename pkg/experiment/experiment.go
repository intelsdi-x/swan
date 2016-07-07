package experiment

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	experimentPhase "github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/nu7hatch/gouuid"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"os"
	"path"
	"strconv"
	"time"
)

// Experiment captures the internal data for the Experiment Driver.
type Experiment struct {
	// Human-readable name.
	// TODO(bp): Push that to DB via Snap in tag or using SwanCollector.
	name string
	// Unique ID for Experiment.
	// Pushed to DB via Snap in tag.
	uuid                string
	workingDirectory    string
	phases              []experimentPhase.Phase
	startingDirectory   string
	experimentDirectory string

	logFile  *os.File
	logLevel log.Level

	textUI bool
}

// NewExperiment creates a new Experiment instance,
// initialize experiment working directory and initialize logs.
// Caller have to provide slice of Phase interfaces which are going to be executed.
func NewExperiment(name string, phases []experimentPhase.Phase,
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
		name:             name,
		uuid:             uuid.String(),
		workingDirectory: directory,
		phases:           phases,
		logLevel:         logLevel,
		textUI:           logLevel == log.ErrorLevel,
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

	log.Info("Starting Experiment ", e.name, " with uuid ", e.uuid)
	// Print uuid on stdout to be able to draw sensitivity profile automatically.
	fmt.Println(e.uuid)

	// Adds progress bar and some brief output when experiment is run in non-verbose
	// mode.
	var bar *pb.ProgressBar
	var increment int
	if e.textUI {
		fmt.Printf("Experiment '%s' with uuid '%s'\n", e.name, e.uuid)
		bar = pb.StartNew(100)
		bar.ShowCounters = false
		bar.ShowTimeLeft = true
		totalPhases := 0
		for _, phase := range e.phases {
			totalPhases += phase.Repetitions()
		}
		increment = 100 / totalPhases
	}

	for id, phase := range e.phases {
		if e.textUI {
			prefix := fmt.Sprintf("[%02d / %02d] %s ", id, len(e.phases), phase.Name())
			bar.Prefix(prefix)
		}

		for i := 0; i < phase.Repetitions(); i++ {
			// Create phase session.
			session := experimentPhase.Session{
				PhaseID:      phase.Name(),
				ExperimentID: e.uuid,
				RepetitionID: i,
			}

			// Start timer.
			phaseStartingTime := time.Now()
			log.Info("Starting ", session.PhaseID, " repetition ", session.RepetitionID)

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

			if e.textUI {
				bar.Add(increment)
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

	if e.textUI {
		bar.Finish()
	}

	log.Info("Ended experiment ", e.name, " with uuid ", e.uuid,
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

	e.experimentDirectory = path.Join(
		e.startingDirectory,
		e.name,
		e.uuid)

	err := os.MkdirAll(e.experimentDirectory, 0777)
	if err != nil {
		return err
	}

	err = os.Chdir(e.experimentDirectory)
	return err
}

// Finalize closes log file and returns to the previous working directory.
func (e *Experiment) Finalize() {
	// Exit experiment directory
	os.Chdir(e.startingDirectory)
}

func (e *Experiment) createPhaseDir(session experimentPhase.Session) error {
	phaseDir := path.Join(e.experimentDirectory,
		session.PhaseID, strconv.FormatInt(int64(session.RepetitionID), 10))

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
	log.SetOutput(io.MultiWriter(e.logFile, os.Stderr))

	return err
}
