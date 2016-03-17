package experiment

// Series of phases: LC workload and combination of isolation, aggressor’s
//workload. Includes baselining when there is no aggressor.
type Experiment struct {
	//Experiment Configuration
	Conf Configuration
	// Shall we make it part of Phases?
	Baseline Phase
	// Phases
	Phases []Phase
	// Target Questions per Seconds
	TargetQPS int
	// Slice with values of the Load Points
	LoadPoints []LoadPoint
}

func (e Experiment) String() string {
	return "Experiment object not defined"
}

// WIP
func tune(exp *Experiment) int {
	return 0
}

func RunExperiment(exp *Experiment) *SensitivityProfile {

	// 1. Configure?

	// 2. Tune
	err := tune(exp)

	for i, _ := range exp.Phases {
		RunPhase(exp, i)
	}
}
