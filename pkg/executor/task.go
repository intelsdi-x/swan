package executor

// TaskState is an enum presenting current task state.
type TaskState int

const (
	// RUNNING task state means that task is still running.
	RUNNING TaskState = iota
	// TERMINATED task state means that task completed or stopped.
	TERMINATED
)

// Task represents a process which can be stopped or monitored.
type Task interface{
	// Stops a task.
	Stop() error
	// Status returns a state of the task. If task is terminated it returns the Status as a
	// second item in tuple. Otherwise returns nil.
	Status() (TaskState, *Status)
	// Waits for the task completion.
	// In case of 0 timeout it will be endlessly blocked.
	// Returns false after timeout exceeds.
	Wait(timeoutMs int) bool
}
