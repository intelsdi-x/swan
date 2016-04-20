package executor

import (
	"time"
)

// TaskState is an enum presenting current task state.
type TaskState int

const (
	// RUNNING task state means that task is still running.
	RUNNING TaskState = iota
	// TERMINATED task state means that task completed or stopped.
	TERMINATED
)

// WaitWithTimeout is a helper for waiting with a given timeout time.
// It returns true if task was NOT timeouted.
func WaitWithTimeout(task Task, timeout time.Duration) (result bool) {
	result = false

	waitErrChannel := make(chan error, 1)
	go func() {
		x := task.Wait()
		waitErrChannel <- x
	}()

	select {
	case err := <-waitErrChannel:
		if err != nil {
			result = true
		}
	case <-time.After(timeout):
	}

	return result
}

// Task represents a process which can be stopped or monitored.
type Task interface {
	// Stops a task.
	Stop() error
	// Status returns a state of the task. If task is terminated it returns the Status as a
	// second item in tuple. Otherwise returns nil.
	Status() (TaskState, *Status)
	// Wait does the blocking wait for the task completion.
	Wait() error
}
