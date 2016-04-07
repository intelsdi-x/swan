package provisioning

// Task represents a process which can be stoped or monitored.
type Task interface{
	// Stops a task.
	Stop()
	// Fetches status of the task
	Status() Status
	// Waits for the task completion.
	// In case of 0 timeout it will be endlessly blocked.
	// Returns true after timeout exceeds.
	Wait(timeoutMs int) bool
}
