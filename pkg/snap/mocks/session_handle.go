package mocks

import "github.com/stretchr/testify/mock"

// SessionHandle is a mock of snap.SessionHandle
type SessionHandle struct {
	mock.Mock
}

// IsRunning provides a mock function with given fields:
func (_m *SessionHandle) IsRunning() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// StopAndRemove provides a mock function with given fields:
func (_m *SessionHandle) StopAndRemove() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WaitForSuccessfulHit provides a mock function with given fields:
func (_m *SessionHandle) WaitForSuccessfulHit() {
	_m.Called()
}
