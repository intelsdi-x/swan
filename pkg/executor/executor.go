package executor

// Executor is responsible for creating execution environment for given workload.
// It is asynchronous.
// It returns a Task interface.
type Executor interface {
	Execute(command string) (Task, error)
}
