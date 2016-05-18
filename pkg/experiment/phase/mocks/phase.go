package mocks

import "github.com/intelsdi-x/swan/pkg/experiment/phase"
import "github.com/stretchr/testify/mock"

// Phase is an autogenerated mock type for the Phase type
type Phase struct {
	mock.Mock
}

// Finalize provides a mock function with given fields:
func (_m *Phase) Finalize() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Name provides a mock function with given fields:
func (_m *Phase) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Repetitions provides a mock function with given fields:
func (_m *Phase) Repetitions() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Run provides a mock function with given fields: _a0
func (_m *Phase) Run(_a0 phase.Session) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(phase.Session) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
