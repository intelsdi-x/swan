package uploaders

import "time"

// Experiment is a gocassa model for experiment metadata
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

// Phase is a gocassa model for phase metadata
type Phase struct {
	ID                  string
	ExperimentID        string
	LCParameters        string
	LCIsolation         string
	AggressorNames      []string
	AggressorParameters []string
	AggressorIsolations []string
}

// Measurement is a gocassa model for phase metadata
type Measurement struct {
	PhaseID      string
	ExperimentID string
	Load         float64
	LoadPointQPS float64
	LGParameters []string
}
