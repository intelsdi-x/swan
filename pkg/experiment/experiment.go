package experiment

import "errors"

type ExperimentConfiguration struct {
	MaxVariance float64
}

type Experiment struct {
	Session Session
	conf    ExperimentConfiguration
	phases  []Phase
	results map[string][]float64
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

	for _, phase := range e.phases {
		//Phase workdir is e.Session.Name + phase.Name()
		for i := 0; i < phase.Repetitions(); i++ {
			result, err := phase.Run()
			if err != nil {
				return err
			}
			e.results[phase.Name()] = append(e.results[phase.Name()], result)
		}

		if variance(e.results[phase.Name()]) > e.conf.MaxVariance {
			return errors.New("Phase max variance exceeded")
		}
	}
	return err
}
