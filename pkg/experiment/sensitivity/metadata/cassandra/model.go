package cassandra

import "github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"

// Experiment is a gocassa model for experiment metadata (such that are constant for all the phases) of sensitivity profile experiment.
type Experiment struct {
	// The tag is needed for embedding; see: https://github.com/hailocab/gocassa#encodingdecoding-data-structures.
	metadata.BaseExperiment `cql:",squash"`
}

// Phase is a gocassa model for phase metadata (such that are constant for all measurements) of sensitivity profile experiment.
type Phase struct {
	// The tag is needed for embedding; see: https://github.com/hailocab/gocassa#encodingdecoding-data-structures.
	metadata.BasePhase  `cql:",squash"`
	ExperimentID        string
	AggressorNames      []string
	AggressorParameters []string
	AggressorIsolations []string
}

// Measurement is a gocassa model for measurement metadata (such that are different for each measurement) of sensitivity profile experiment.
type Measurement struct {
	PhaseID      string
	ExperimentID string
	// The tag is needed for embedding; see: https://github.com/hailocab/gocassa#encodingdecoding-data-structures.
	metadata.BaseMeasurement `cql:",squash"`
}
