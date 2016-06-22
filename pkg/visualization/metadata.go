package visualization

// ExperimentMetadata encodes the metadata which is related to an experiment run.
// This currently only contains the experiment id, but is intended to encode
// the experiment environment (hardware and software configuration),
// the machines involved in the experiment, etc.
type ExperimentMetadata struct {
	experimentID string
}

// NewExperimentMetadata is the ExperimentMetadata constructor and returns
// a new ExperimentMetadata with a specific id.
func NewExperimentMetadata(ID string) *ExperimentMetadata {
	return &ExperimentMetadata{
		ID,
	}
}

// String returns a printable string with all experiment metadata.
// This is currently only the experiment id.
func (metadata *ExperimentMetadata) String() string {
	return "Experiment id: " + metadata.experimentID
}
