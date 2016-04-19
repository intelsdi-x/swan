package experiment

import (
	"github.com/nu7hatch/gouuid"
)

// Experiment holds enough metadata to execute each step (phase) which defines
// the experiment. Further more, the experiment stores artifact files
// (output files from measurements etc) according to the name of the experiment
// and phases.
type Experiment struct {
	Name   string
	Phases Phases
}

// NewExperiment constructs a new Experiment with a name and a list of phases.
func NewExperiment(name string, phases Phases) *Experiment {
	return &Experiment{
		Name:   name,
		Phases: phases,
	}
}

// AddPhase adds a phase to the experiment. Phases must implement the phase
// interface defined in phase.go.
func (experiment *Experiment) AddPhase(phase Phase) {
	experiment.Phases = append(experiment.Phases, phase)
}

// Run executes each phase in the experiment a configurable number of times
// (defined by the Repetitions() implementation in each phase).
func (experiment *Experiment) Run() (*Session, error) {
	// Generate session name
	sessionName, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	session := &Session{Name: sessionName.String()}

	for _, phase := range experiment.Phases {
		repetitions := phase.Repetitions()

		for repetition := 0; repetition < repetitions; repetition++ {
			_, err = phase.Run()

			// Abort experiment if one phase failed
			if err != nil {
				return nil, err
			}
		}
	}

	return session, nil
}
