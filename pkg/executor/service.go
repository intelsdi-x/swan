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

package executor

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

/**
ServiceLauncher and ServiceHandle are wrappers that could be used on Launcher and TaskHandle class.
User should use them to state intent that these processes should not stop without
explicit `Stop()` or `Wait()` invoked on TaskHandle.

If process would stop on it's own, the Stop() and Wait() functions will return error
and process logs will be available on experiment log stream. In this case, each subsequent invocation of Stop() and Wait() will return error.
*/

// ServiceLauncher is a decorator and Launcher implementation that should be used for tasks that do not stop on their own.
type ServiceLauncher struct {
	Launcher
}

// Launch implements Launcher interface.
func (sl ServiceLauncher) Launch() (TaskHandle, error) {
	th, err := sl.Launcher.Launch()
	if err != nil {
		return nil, err
	}

	return NewServiceHandle(th), nil
}

type serviceHandle struct {
	TaskHandle
	taskHasBeenTerminatedByUser bool
	err                         error
	mutex                       *sync.Mutex
}

// NewServiceHandle wraps TaskHandle with serviceHandle.
func NewServiceHandle(handle TaskHandle) TaskHandle {
	return &serviceHandle{
		TaskHandle: handle,
		mutex:      &sync.Mutex{}}
}

// Stop implements TaskHandle interface.
func (s *serviceHandle) Stop() error {
	err := s.checkErrorCondition()
	if err != nil {
		return err
	}

	return s.TaskHandle.Stop()
}

// Wait implements TaskHandle interface.
func (s *serviceHandle) Wait(duration time.Duration) bool {
	return s.TaskHandle.Wait(duration)
}

func (s *serviceHandle) checkErrorCondition() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// When ServiceHandle encountered error, same error will be returned every time.
	if s.err != nil {
		return s.err
	}

	// When task has been Stopped by user, then no error is returned when task is terminated.
	if !s.taskHasBeenTerminatedByUser {
		s.taskHasBeenTerminatedByUser = true
		if s.TaskHandle.Status() == TERMINATED {
			s.err = errors.Errorf("ServiceHandle with command %q has terminated prematurely", s.TaskHandle.Name())
			logrus.Errorf(s.err.Error())
			logOutput(s.TaskHandle)
			return s.err
		}
	}
	return nil
}
