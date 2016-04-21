package experiment

import (
	"errors"
	"log"
	"os"
)

// InputConfiguration structure holds input variable
// that instruments Experiment Driver.
// MaxVariance - maximum allowed variance of result between
//               different phase runs.
// WorkingDirectory - path where Experiment Driver shall put all experiment
//               logs.
type InputConfiguration struct {
	MaxVariance      float64
	WorkingDirectory string
}

// Experiment structure internal data for the Experiment Driver
type Experiment struct {
	session             session
	conf                InputConfiguration
	phases              []Phase
	results             map[string][]float64
	startingDirectory   string
	experimentDirectory string
	logFile             *os.File
	logger              *log.Logger
}

// NewExperiment function creates new Experiment instance,
// initialize experiment working directory and initialize
// logs.
// User shall provide slice of Phase interfaces which are
// going to be executed.
func NewExperiment(
	configuration InputConfiguration,
	phases []Phase) (*Experiment, error) {

	if configuration.MaxVariance <= 0 {
		return nil, errors.New("Invalid argument: variance")
	}
	if len(phases) == 0 {
		return nil, errors.New("Invalid argument: nil phase slice")
	}

	// mpatelcz TODO: Check if phase names are unique!

	session := sessionNew()
	e := &Experiment{
		session: session,
		conf:    configuration,
		phases:  phases,
		results: make(map[string][]float64, len(phases)),
	}
	err := e.mkExperimentDir()
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
// When phase repetition is over results variance is checked to see
// if results are in given range.
// When variance exceeds given limit experiment is stopped.
func (e *Experiment) Run() error {
	var err error

	// Always perform cleanup on exit.
	defer e.finalize()

	for _, phase := range e.phases {
		for i := 0; i < phase.Repetitions(); i++ {
			e.logger.Printf("Starting Phase: '%s', iteration: %d\n",
				phase.Name(), i)
			err = e.mkPhaseDir(phase, i)
			if err != nil {
				return err
			}
			result, err := phase.Run()
			if err != nil {
				e.logger.Print("Phase returned error ", err)
				return err
			}
			e.logger.Printf("Phase ended with success. Returned: %2.4f\n", result)
			e.results[phase.Name()] = append(e.results[phase.Name()], result)
		}

		variance := variance(e.results[phase.Name()])
		e.logger.Printf("Phase repetitions variance %2.4f\n", variance)

		if variance > e.conf.MaxVariance {
			e.logger.Printf("Variance %2.4f exceeded limit %2.4f. Exiting\n",
				variance, e.conf.MaxVariance)
			return errors.New("Phase max variance exceeded")
		}
		e.logger.Printf("Phase variance %2.4f is within limit %2.4f\n",
			variance, e.conf.MaxVariance)
	}
	e.logger.Println("Done with measurement.")
	return err
}
