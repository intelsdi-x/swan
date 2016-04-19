package mocks

import "github.com/stretchr/testify/mock"

// LoadGenerator mock
type LoadGenerator struct {
	mock.Mock
}

// Load provides a mock function with given fields: qps, durationMs
func (_m *LoadGenerator) Load(qps int, durationMs int) (int, error) {
	ret := _m.Called(qps, durationMs)

	var r0 int
	if rf, ok := ret.Get(0).(func(int, int) int); ok {
		r0 = rf(qps, durationMs)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int, int) error); ok {
		r1 = rf(qps, durationMs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Tune provides a mock function with given fields: slo, timeoutMs
func (_m *LoadGenerator) Tune(slo int, timeoutMs int) (int, error) {
	ret := _m.Called(slo, timeoutMs)

	var r0 int
	if rf, ok := ret.Get(0).(func(int, int) int); ok {
		r0 = rf(slo, timeoutMs)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int, int) error); ok {
		r1 = rf(slo, timeoutMs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
