package sensitivity

import "github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"

// Uploader is a interface for SwanMetrics uploading into external media(like database or telemetry framework)
type Uploader interface {
	SendMetadata(metadata.Experiment) error
	GetMetadata(experiment string) (metadata.Experiment, error)
}
