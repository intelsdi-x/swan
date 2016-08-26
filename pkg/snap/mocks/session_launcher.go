package mocks

import "github.com/intelsdi-x/athena/pkg/snap"
import "github.com/stretchr/testify/mock"

import "github.com/intelsdi-x/athena/pkg/executor"

type SessionLauncher struct {
	mock.Mock
}

// LaunchSession provides a mock function with given fields: _a0, _a1
func (_m *SessionLauncher) LaunchSession(_a0 executor.TaskInfo, _a1 string) (snap.SessionHandle, error) {
	ret := _m.Called(_a0, _a1)

	var r0 snap.SessionHandle
	if rf, ok := ret.Get(0).(func(executor.TaskInfo, string) snap.SessionHandle); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(snap.SessionHandle)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(executor.TaskInfo, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
