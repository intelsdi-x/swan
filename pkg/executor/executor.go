package executor

// Executor is responsible for creating execution environment for given workload.
// It returns Task handle when workload started gracefully.
// Workload is executed asynchronously.
type Executor interface {
	// Execute executes command on underlying platform.
	// Invokes "bash -c <command>" and waits for short time to make sure that process has started.
	// Returns error if command exited immediately with non-zero exit status.
	Execute(command string) (TaskHandle, error)
	// Name returns user-friendly name of executor.
	Name() string
}
