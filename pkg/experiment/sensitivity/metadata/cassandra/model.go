package cassandra

import "github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"

// Experiment is a gocassa model for experiment metadata (such that are constant for all the phases) of sensitivity profile experiment.
type Experiment struct {
	metadata.BaseExperiment
}

// Phase is a gocassa model for phase metadata (such that are constant for all measurements) of sensitivity profile experiment.
type Phase struct {
	metadata.BasePhase
	ExperimentID        string
	AggressorNames      []string
	AggressorParameters []string
	AggressorIsolations []string
}

// Measurement is a gocassa model for measurement metadata (such that are different for each measurement) of sensitivity profile experiment.
type Measurement struct {
	PhaseID      string
	ExperimentID string
	metadata.Measurement
}
