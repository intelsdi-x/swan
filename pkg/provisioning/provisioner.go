package provisioning

// Provisioner is responsible for creating execution environment for given
// workload with given isolation. It returns a pointer to the Task.
// TODO(bp): Decide about the name: Provisioning vs Runner vs ...
type Provisioner interface{
	Run(command string) (Task, error)
}
