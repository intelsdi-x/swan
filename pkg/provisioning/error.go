package provisioning

import "fmt"

type ProvisioningError struct {
	s string
}

func (e *ProvisioningError) Error() string {
	return e.s
}

func NewError(args... interface{}) error {
	return &ProvisioningError{fmt.Sprint(args...)}
}
