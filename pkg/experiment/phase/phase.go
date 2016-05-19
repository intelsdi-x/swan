package phase

// Phase defines interface which shall be provided by user for the
// Experiment Driver.
type Phase interface {
	// Name returns measurement name.
	Name() string
	// Repetitions returns desired number of measurement repetitions.
	Repetitions() int
	// Run runs a measurement. It takes phase session to make each phase
	// unique for collected results.
	Run(Session) error
	// Finalize is executed after all repetitions of given measurement.
	Finalize() error
}
