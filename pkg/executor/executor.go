package executor

// Executor is responsible for creating execution environment for given workload.
// Task is executed asynchronously then and task handle is returned.
type Executor interface {
	Execute(command string) (Task, error)
}
