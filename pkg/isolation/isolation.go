package isolation

// Isolation abstraction gives ability for different implementation to set isolation before
// and after task start.
type Isolation interface{
	// Method executed after task started.
	Isolate(taskPid TaskPID) error
}
