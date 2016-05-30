package metrics

// Uploader is a interface for SwanMetrics uploading into external media(like database or telemetry framework)
type Uploader interface {
	SendMetrics(Swan) error
}
