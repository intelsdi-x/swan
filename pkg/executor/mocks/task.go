package mocks

import "github.com/intelsdi-x/swan/pkg/executor"
import "github.com/stretchr/testify/mock"

// Task mock
type Task struct {
	mock.Mock
}

// Status provides a mock function with given fields:
func (_m *Task) Status() (executor.TaskState, *executor.Status) {
	ret := _m.Called()

	var r0 executor.TaskState
	if rf, ok := ret.Get(0).(func() executor.TaskState); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(executor.TaskState)
	}

	var r1 *executor.Status
	if rf, ok := ret.Get(1).(func() *executor.Status); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*executor.Status)
		}
	}

	return r0, r1
}

// Stop provides a mock function with given fields:
func (_m *Task) Stop() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Wait provides a mock function with given fields: timeoutMs
func (_m *Task) Wait() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
