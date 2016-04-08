package provisioning

// Status represents the status of a task in the current point of time.
// NOTE: We need to define if we user wants status and output in one struct.
type Status struct {
	code   int
	stdout string
	stderr string
}
