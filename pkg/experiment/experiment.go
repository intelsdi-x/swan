package experiment

import (
	"errors"
	"math"
)

type PhaseFunc func() (float64, error)

type Configuration struct {
	MaxVariance   float64
	PhaseRepCount int
}

type Experiment struct {
	conf    Configuration
	phases  []PhaseFunc
	results []float64
}

// Construct new Experiment object.
func NewExperiment(
	configuration Configuration,
	phases []PhaseFunc) *Experiment {

	return &Experiment{
		conf:   configuration,
		phases: phases,
	}
}

func (e *Experiment) AddPhase(p PhaseFunc) error {
	e.phases = append(e.phases, p)
	return nil
}

func average(x []float64) float64 {
	var avr float64
	for _, val := range x {
		avr += val
	}
	avr /= float64(len(x))
	return avr
}
func variance(x []float64) float64 {

	avr := average(x)
	variance := float64(0)
	for _, val := range x {
		variance += math.Sqrt(math.Abs(avr - val))
	}
	variance /= float64(len(x))
	return variance
}

func (e *Experiment) Run() error {
	var err error

	for _, phase := range e.phases {
		for i := 0; i < e.conf.PhaseRepCount; i++ {
			result, err := phase()
			if err != nil {
				return nil
			}
			e.results = append(e.results, result)
		}
		if variance(e.results) > e.conf.MaxVariance {
			return errors.New("Phase max variance exceeded")
		}
	}
	return err
}
