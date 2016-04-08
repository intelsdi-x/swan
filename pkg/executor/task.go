package executor

// Task represents a process which can be stoped or monitored.
type Task interface{
	// Stops a task.
	Stop() error
	// Fetches status of the task when the task is terminated. Otherwise return running true
	// and nil
	Status() (bool, *Status)
	// Waits for the task completion.
	// In case of 0 timeout it will be endlessly blocked.
	// Returns false after timeout exceeds.
	Wait(timeoutMs int) bool
}
