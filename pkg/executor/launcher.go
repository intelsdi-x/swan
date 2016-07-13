package executor

// Launcher responsibility is to launch previously configured job.
type Launcher interface {
	// Launch starts the workload (process or group of processes). It returns a workload
	// represented as a Task Handle instance.
	// Error is returned when Launcher is unable to start a job.
	Launch() (TaskHandle, error)

	// Name returns human readable name for job.
	// TODO(bp): Do the same for LoadGenerator.
	Name() string

	// TODO(bp): Include a getter for parameters (part of SCE-376).
}
