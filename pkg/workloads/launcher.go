package workloads

// Launcher defines how workload should be run using specified Provisioners.
type Launcher interface {
	// Launch launches the workload.
	// Returned Task is able to stop & monitor the provisioned process.
	Launch() (Task, error)
}
