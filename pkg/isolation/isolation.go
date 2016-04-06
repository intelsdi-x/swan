package isolation

// Isolation abstraction gives ability for different implementation to set isolation before
// and after task start.
type Isolation interface{
	// Method executed before task start.
	Init() error
	// Method executed after task started.
	Perform(taskPid TaskPID) error
}
