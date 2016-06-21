package visualization

import (
	"fmt"
)

// ExperimentMetadata is a model for data.
type ExperimentMetadata struct {
	experimentID string
}

// NewExperimentMetadata creates new model of data representation.
func NewExperimentMetadata(ID string) *ExperimentMetadata {
	return &ExperimentMetadata{
		ID,
	}
}

// PrintExperimentMetadata prints elements from list.
func PrintExperimentMetadata(experimentMetadata *ExperimentMetadata) {
	fmt.Println("\nExperiment id: " + experimentMetadata.experimentID)
}
