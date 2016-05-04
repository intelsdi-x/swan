package mocks

import "github.com/intelsdi-x/swan/pkg/executor"
import "github.com/stretchr/testify/mock"

import "io"
import "time"

// Task is an autogenerated mock type for the Task type
type Task struct {
	mock.Mock
}

// Clean provides a mock function with given fields:
func (_m *Task) Clean() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EraseOutput provides a mock function with given fields:
func (_m *Task) EraseOutput() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Status provides a mock function with given fields:
func (_m *Task) Status() (executor.TaskState, int) {
	ret := _m.Called()

	var r0 executor.TaskState
	if rf, ok := ret.Get(0).(func() executor.TaskState); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(executor.TaskState)
	}

	var r1 int
	if rf, ok := ret.Get(1).(func() int); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(int)
	}

	return r0, r1
}

// Stderr provides a mock function with given fields:
func (_m *Task) Stderr() (io.Reader, error) {
	ret := _m.Called()

	var r0 io.Reader
	if rf, ok := ret.Get(0).(func() io.Reader); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
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

// Stdout provides a mock function with given fields:
func (_m *Task) Stdout() (io.Reader, error) {
	ret := _m.Called()

	var r0 io.Reader
	if rf, ok := ret.Get(0).(func() io.Reader); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
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

// Wait provides a mock function with given fields: timeout
func (_m *Task) Wait(timeout time.Duration) bool {
	ret := _m.Called(timeout)

	var r0 bool
	if rf, ok := ret.Get(0).(func(time.Duration) bool); ok {
		r0 = rf(timeout)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
