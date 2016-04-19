package experiment

import "github.com/intelsdi-x/swan/pkg/executor"

// Phase responsibility is to launch previously configured job.
type Phase interface {
	Name() string

	// Launch starts the workload (process or group of processes). It returns a workload
	// represented as a Task instance.
	// Error is returned when Launcher is unable to start a job.
	Run() (*executor.Task, error)

	Repetitions() int
}

// Phases is a convenience type which defines a set of phases.
// Useful when creating an experiment without knowing the needed phases ahead
// of time.
type Phases []Phase
