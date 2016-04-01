package experiment

// Measurememt
//
// Getting SLI for LC workload at given load point for given experimental
// conditions
type Measurement struct {
	//
	//
	//
	// Result from measurement
	sli SLI
}

func (m Measurement) String() string {
	return "Measurement object not defined"
}

func RunMeasurement(exp *Experiment, wl *Workload, lp *LoadPoint) int {

	// Here is the place for running commands
	//1 Prepare environment? (something from Experiment?)
	//   - prepare for the measurement

	sli, err := wl.Run(lp)

	//For what I need Experiment?

	// Return success
	return 0
}
