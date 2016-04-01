package experiment

// Measurements for a specific aggressor (or none) at different load points for
// given experimental conditions
type Phase struct {
	//
	workload Workload
	//
	isolation Isolation
	// Series of measurements which defines this Phase
	measuremetns []Measurement
}

func (p Phase) String() string {
	return "Phase object not defined"
}

func phaseSetIsolation() int {
	return 0
}

func RunPhase(exp *Experiment, ph *Phase) int {

	//What to do with isolation?
	phaseSetIsolation()

	for i, _ := range ph.measuremetns {
		for _, lp := range exp.LoadPoints {
			RunMeasurement(exp, &ph.workload, &lp)
		}
	}
	return 0
}
