package experiment

import "errors"

type PhaseFunc func() (float64, error)

type Phase struct {
	Name      string
	PhaseFunc PhaseFunc
}

type ExperimentConfiguration struct {
	MaxVariance   float64
	PhaseRepCount int
}

type Experiment struct {
	conf    ExperimentConfiguration
	phases  []Phase
	results []float64
}

// Construct new Experiment object.
func NewExperiment(
	configuration ExperimentConfiguration,
	phases []Phase) (*Experiment, error) {

	if configuration.MaxVariance <= 0 {
		return nil, errors.New("Invalid argument: variance")
	}
	if configuration.PhaseRepCount <= 0 {
		return nil, errors.New("Invalid argument: PhaseRepCount")
	}
	if len(phases) == 0 {
		return nil, errors.New("Invalid argument: nil phase slice")
	}
	for _, phase := range phases {
		if phase.PhaseFunc == nil {
			return nil, errors.New("Invalid argument: nil PhaseFunc")
		}
	}

	return &Experiment{
		conf:   configuration,
		phases: phases,
	}, nil
}

func (e *Experiment) Run() error {
	var err error

	for _, phase := range e.phases {
		for i := 0; i < e.conf.PhaseRepCount; i++ {
			result, err := phase.PhaseFunc()
			if err != nil {
				return err
			}
			e.results = append(e.results, result)
		}
		if variance(e.results) > e.conf.MaxVariance {
			return errors.New("Phase max variance exceeded")
		}
	}
	return err
}
