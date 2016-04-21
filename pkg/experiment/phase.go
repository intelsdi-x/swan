package experiment

// Phase defines interface which shall be provided by user for the
// Experiment Driver.
// It shall have following function:
// Name() - which will return phase name.
// Repetitions() - which will return desired number of phase repetitions for
//          calculating variance.
// Run() - which will run a phase measurements.
type Phase interface {
	Name() string
	Run() (float64, error)
	Repetitions() int
}
