package executor

// Executor is responsible for creating execution environment for given workload.
// It returns Task handle when workload started gracefully.
// Workload is executed asynchronously.
type Executor interface {
	// Execute executes command on underlying platform.
	Execute(command string) (TaskHandle, error)
	// Name returns user-friendly name of executor.
	Name() string
}
