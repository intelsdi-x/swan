package workloads

import "github.com/intelsdi-x/swan/pkg/executor"

// Launcher responsibility is to launch previously configured job.
type Launcher interface {
	// Launch starts the workload (process or group of processes). It returns a workload
	// represented as a Task Handle instance.
	// Error is returned when Launcher is unable to start a job.
	Launch() (executor.TaskHandle, error)
}
