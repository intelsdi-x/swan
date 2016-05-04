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
	// second item in tuple. Otherwise returns -1.
	Status() (TaskState, int)
	// Stdout returns a reader for file to the task's stdout file.
	Stdout() (io.Reader, error)
	// Stderr returns a reader for file to the task's stderr file.
	Stderr() (io.Reader, error)
	// Wait does the blocking wait for the task completion in case of nil.
	// Wait is a helper for waiting with a given timeout time.
	// It returns true if task is terminated.
	Wait(timeout time.Duration) bool
	// Clean cleans task temporary resources like isolations for Local.
	// It also closes the task's stdout & stderr files.
	Clean() error
	// EraseOutput removes task's stdout & stderr files.
	EraseOutput() error
}
