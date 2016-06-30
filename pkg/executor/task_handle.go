package executor

import (
	"os"
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

// TaskInfo represents process from which information like Stdout can be retrieved.
// NOTE: In future it can include other info like `Address`.
type TaskInfo interface {
	// StdoutFile returns a file handle for file to the task's stdout file.
	StdoutFile() (*os.File, error)
	// StderrFile returns a file handle for file to the task's stderr file.
	StderrFile() (*os.File, error)
}

// TaskHandle represents a process which can be stopped or monitored.
type TaskHandle interface {
	TaskInfo
	// Stops a task.
	Stop() error
	// Status returns a state of the task.
	Status() TaskState
	// ExitCode returns a exitCode. If task is not terminated it returns error.
	ExitCode() (int, error)
	// Wait does the blocking wait. It is a helper for waiting with a given timeout time.
	// It returns true if task is terminated.
	// In case of '0*time.Nanoseconds' timeout it waits infinitely for task completion.
	Wait(timeout time.Duration) bool
	// Clean cleans task temporary resources like isolations for Local.
	// It also closes the task's stdout & stderr files.
	Clean() error
	// EraseOutput removes task's stdout & stderr files.
	EraseOutput() error
	// Location returns address where task was located.
	Address() string
}
