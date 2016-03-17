package experiment

// Measurements for a specific aggressor (or none) at different load points for
// given experimental conditions
type Phase struct {
	//
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

func RunPhase(exp *Experiment, no int) int {

	//What to do with isolation?
	phaseSetIsolation()

	for i, _ := range exp.phases[no].measurements {
		RunMeasurement(exp)
	}
	return 0
}
