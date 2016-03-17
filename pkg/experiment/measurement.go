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

func RunMeasurement(exp *Experiment, lp *LoadPoint) int {

	// Here is the place for running commands
	return 0
}
