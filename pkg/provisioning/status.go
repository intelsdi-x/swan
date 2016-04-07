package provisioning

const (
	// SuccessCode is a successful status code.
	SuccessCode = 0
	// RunningCode points out that the Process is still running.
	RunningCode = 9999
)


// Status represents the status for a command.
type Status struct {
	code int
 	stdout string
	stderr string
}
