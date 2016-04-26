package executor

import (
	"io"
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

// Task represents a process which can be stopped or monitored.
type Task interface {
	// Stops a task.
	Stop() error
	// Status returns a state of the task. If task is terminated it returns exitCode as a
	// second item in tuple. Otherwise returns nil.
	Status() (TaskState, *Status)
	// Stdout returns reader for file to which current task was writing stdout.
	// If file is removed returns error.
	Stdout() (io.Reader, error)
	// Stderr returns reader for file to which current task was writing stderr.
	// If file is removed returns error.
	Stderr() (io.Reader, error)
	// Wait does the blocking wait for the task completion in case of nil.
	// Wait is a helper for waiting with a given timeout time.
	// It returns true if task is terminated.
	Wait(timeout time.Duration) bool
	// Clean cleans task temporary resources.
	// For Local & Remote Task it removes files to which stdout and stderr
	// of current task was written.
	Clean() error
}
