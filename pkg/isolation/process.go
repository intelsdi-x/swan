package isolation

// TaskPID is a Linux Process ID.
type TaskPID int64

// ProcessIsolation abstraction gives ability for different implementation to set isolation before
// and after task start.
type ProcessIsolation interface{
	// Method executed after task started.
	Isolate(taskPid TaskPID) error
}
