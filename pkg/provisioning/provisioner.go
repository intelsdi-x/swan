package provisioning

import "github.com/intelsdi-x/swan/pkg/isolation"

type Provisioner interface{
	Execute(string, string, []isolation.Isolation) <-chan Status
}
