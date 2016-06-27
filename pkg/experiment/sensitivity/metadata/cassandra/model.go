package cassandra

import "time"

// Experiment is a gocassa model for experiment metadata (such that are constant for all the phases) of sensitivity profile experiment.
type Experiment struct {
	ID                string
	LoadDuration      time.Duration
	TuningDuration    time.Duration
	LcName            string
	LgNames           []string
	RepetitionsNumber int
	LoadPointsNumber  int
	SLO               int
}

// Phase is a gocassa model for phase metadata (such that are constant for all measurements) of sensitivity profile experiment.
type Phase struct {
	ID                  string
	ExperimentID        string
	LCParameters        string
	LCIsolation         string
	AggressorNames      []string
	AggressorParameters []string
	AggressorIsolations []string
}

// Measurement is a gocassa model for measurement metadata (such that are different for each measurement) of sensitivity profile experiment.
type Measurement struct {
	PhaseID      string
	ExperimentID string
	Load         float64
	LoadPointQPS float64
	LGParameters []string
}
