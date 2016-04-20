package mocks

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/stretchr/testify/mock"
)

// Launcher mock
type Launcher struct {
	mock.Mock
}

// Launch provides a mock function with given fields:
func (_m *Launcher) Launch() (executor.Task, error) {
	ret := _m.Called()

	var r0 executor.Task
	if rf, ok := ret.Get(0).(func() executor.Task); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(executor.Task)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
