package experiment

import (
	"errors"
	"log"
	"os"
)

type ExperimentConfiguration struct {
	MaxVariance      float64
	WorkingDirectory string
}

type Experiment struct {
	Session             Session
	conf                ExperimentConfiguration
	phases              []Phase
	results             map[string][]float64
	startingDirectory   string
	experimentDirectory string
	logFile             *os.File
	logger              *log.Logger
}

// Construct new Experiment object.
func NewExperiment(
	configuration ExperimentConfiguration,
	phases []Phase) (*Experiment, error) {

	if configuration.MaxVariance <= 0 {
		return nil, errors.New("Invalid argument: variance")
	}
	if len(phases) == 0 {
		return nil, errors.New("Invalid argument: nil phase slice")
	}

	// mpatelcz TODO: Check if phase names are unique!

	session := sessionNew()
	return &Experiment{
		Session: session,
		conf:    configuration,
		phases:  phases,
		results: make(map[string][]float64, len(phases)),
	}, nil
}

func (e *Experiment) Run() error {
	var err error

	e.logInitialize()
	for _, phase := range e.phases {
		e.logger.Print("Phase: " + phase.Name())
		for i := 0; i < phase.Repetitions(); i++ {
			err = e.logMkPhase(phase.Name(), i)
			if err != nil {
				e.logClose()
				return err
			}
			e.logger.Printf("   Repetition %v\n", i)
			result, err := phase.Run()
			if err != nil {
				e.logger.Print("Phase returned error ", err)
				e.logClose()
				return err
			}
			e.logger.Print("Phase OK")
			e.results[phase.Name()] = append(e.results[phase.Name()], result)
		}

		variance := variance(e.results[phase.Name()])
		e.logger.Printf("Phase repetitions variance %2.4f\n", variance)
		if variance > e.conf.MaxVariance {
			e.logger.Printf("Variance %2.4f exceeded limit %2.4f. Exiting\n",
				variance, e.conf.MaxVariance)
			e.logClose()
			return errors.New("Phase max variance exceeded")
		}
		e.logger.Print("Phase variance OK")
	}
	e.logger.Println("Done with measurement.")
	e.logClose()
	return err
}
