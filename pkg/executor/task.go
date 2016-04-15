package executor

import "time"

// TaskState is an enum presenting current task state.
type TaskState int

const (
	// RUNNING task state means that task is still running.
	RUNNING TaskState = iota
	// TERMINATED task state means that task completed or stopped.
	TERMINATED
)

// WaitWithTimeout is a helper for waiting with a given timeout time in ms.
// It returns with true if task was NOT timeouted.
// NOTE: It does not mean that it is terminated. E.g. When we do multiple waits.
func WaitWithTimeout(task Task, timeoutMs int) (result bool) {
	result = true
	waitChannel := make(chan bool)

	go func() {
		// Wait will return immediately when the Wait was already triggered.
		task.Wait()
		waitChannel <- true
	}()

	timeoutDuration := time.Duration(timeoutMs) * time.Millisecond

	select {
	case result = <-waitChannel:
	case <-time.After(timeoutDuration):
		result = false
	}

	return
}

// Task represents a process which can be stopped or monitored.
type Task interface {
	// Stops a task.
	Stop() error
	// Status returns a state of the task. If task is terminated it returns the Status as a
	// second item in tuple. Otherwise returns nil.
	Status() (TaskState, *Status)
	// Wait does the blocking wait for the task completion.
	Wait()
	// Cleans the environment after task. E.g Clear temp output files.
	Clean()
}
