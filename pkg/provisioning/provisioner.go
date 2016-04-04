package provisioning

import "github.com/intelsdi-x/swan/pkg/isolation"

// Provisioner is responsible for creating execution evnironment for given
// workload with given isolation. It is always asynchronous and returns Status chan.
type Provisioner interface{
	Execute(string, string, []isolation.Isolation) <-chan Status
}
