package provisioning

const (
	// Success status code.
	SuccessCode = 0
	// Code pointing out that the Process is still running.
	RunningCode = 9999
)


// Status represents the exit status for a command.
type Status struct {
	code int
 	stdout string
	stderr string
}
