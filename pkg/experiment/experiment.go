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

type Experimenter interface {
	//Prepare all
	SetEnvironment() error
	//Find target QPS
	Tune() int
	//Run baseline or workload with aggressor
	RunPhase(no int)
	//Return number of phases
	NumPhases() int
	//Create Phase Isolation
	CreateIsolation() error
	//
	CreateSensitivityProfile() SensitivityProfile
}

type Measurementer interface {
	//Run single measurement
	RunMeasurement(num_phase int, load_point int)
}

// tune - look for the target QPS
func tune(exp *Experiment) int {
	exp.TargetQPS = 0
	return 0
}

// Called after all measurements has been launch.
// Extracts SLI from Measurements and creates
// single sensitivity profile
func createSentivityProfile(exp *Experiment) *SensitivityProfile {
	return &SensitivityProfile{}
}

// Run the main experiment. First Phase shall be Baselining.
func RunExperiment(exp Experimenter) SensitivityProfile {

	exp.SetEnvironment()

	exp.Tune()

	for i := 0; i < exp.NumPhases(); i++ {
		exp.RunPhase(i)
	}

	sp := exp.CreateSensitivityProfile()
	return sp
}
