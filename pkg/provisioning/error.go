package provisioning

import "fmt"

// Error is to compose different error using string message.
type Error struct {
	s string
}

func (e *Error) Error() string {
	return e.s
}

// NewError constructs Swan Error struct.
func NewError(args... interface{}) error {
	return &Error{fmt.Sprint(args...)}
}
