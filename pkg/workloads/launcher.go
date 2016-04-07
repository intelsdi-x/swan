package workloads

// Launcher responsibility is to launch previously configured job.
type Launcher interface {
	// Launch launches the workload.
	// Returned Task is able to stop & monitor the provisioned process.
	Launch() (Task, error)
}
