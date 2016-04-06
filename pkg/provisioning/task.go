package provisioning

// Task represents a process which can be stoped or monitored.
type Task interface{
	// Stops a task. When it cannot be stopped it returns error.
	Stop() error
	// Fetches status of the task
	Status() Status
	// Waits for the task completion. In case of 0 timeout it will be endlessly blocked.
	Wait(timeout int) error
}


