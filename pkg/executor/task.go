package executor

import "io"

// TaskState is an enum presenting current task state.
type TaskState int

const (
	// RUNNING task state means that task is still running.
	RUNNING TaskState = iota
	// TERMINATED task state means that task completed or stopped.
	TERMINATED
)

// Task represents a process which can be stopped or monitored.
type Task interface {
	// Stops a task.
	Stop() error
	// Status returns a state of the task. If task is terminated it returns exitCode as a
	// second item in tuple. Otherwise returns nil.
	Status() (TaskState, *int)
	// Clean removes files to which stdout and stderr of current task was written.
	Clean() error
	// Stdout returns reader for file to which current task was writing stdout.
	Stdout() io.Reader
	// Stderr returns reader for file to which current task was writing stderr.
	Stderr() io.Reader
	// Waits for the task completion.
	// In case of 0 timeout it will be endlessly blocked.
	// Returns false after timeout exceeds.
	Wait(timeoutMs int) bool
}
