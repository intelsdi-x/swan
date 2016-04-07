package provisioning

const (
	// SuccessCode is a successful status code.
	SuccessCode = 0
	// RunningCode points out that the Process is still running.
	RunningCode = 9999
)

// Status represents the status of a task in the current point of time.
// NOTE: We need to define if we user wants status & output in one struct.
// While having
type Status struct {
	code int
 	stdout string
	stderr string
}
