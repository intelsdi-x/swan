package experiment

import "errors"

type PhaseFunc func() (float64, error)

type Configuration struct {
	MaxVariance   int
	PhaseRepCount int
}

type Experiment struct {
	ec     Configuration
	phases []PhaseFunc
	slis   []float64
}

// Construct new Experiment object.
func NewExperiment(configuration Configuration) *Experiment {

	// TODO(mpatelcz): Validate configuration.

	return &Experiment{
		ec: configuration,
	}
}

func (e *Experiment) AddPhase(p PhaseFunc) error {
	e.phases = append(e.phases, p)
	return nil
}

// Prints nice output. TBD
func (e *Experiment) PrintSensitivityProfile() error {
	return nil
}

func (e *Experiment) Run() error {
	var err error
	var min, max float64

	for _, phase := range e.phases {
		for i := 0; i < e.ec.PhaseRepCount; i++ {
			sli, err := phase()
			if sli < min {
				min = sli
			}
			if sli > max {
				max = sli
			}
			if err != nil {
				return nil
			}
			e.slis = append(e.slis, sli)
		}
		if int(max-min) > e.ec.MaxVariance {
			return errors.New("Phase max variance exceeded")
		}
	}

	//That's TBD.
	err = e.PrintSensitivityProfile()
	return err
}
