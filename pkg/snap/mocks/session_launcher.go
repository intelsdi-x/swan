// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mocks

import "github.com/intelsdi-x/swan/pkg/snap"
import "github.com/stretchr/testify/mock"

import "github.com/intelsdi-x/swan/pkg/executor"

// SessionLauncher ...
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
