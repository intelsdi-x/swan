package mocks

import "github.com/intelsdi-x/swan/pkg/executor"
import "github.com/stretchr/testify/mock"

// Executor mock
type Executor struct {
	mock.Mock
}

// Execute provides a mock function with given fields: command
func (_m *Executor) Execute(command string) (executor.Task, error) {
	ret := _m.Called(command)

	var r0 executor.Task
	if rf, ok := ret.Get(0).(func(string) executor.Task); ok {
		r0 = rf(command)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(executor.Task)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(command)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
