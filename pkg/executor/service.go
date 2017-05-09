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
	taskCanBeTerminated bool
	err                 error
	mutex               *sync.Mutex
}

// NewServiceHandle wraps TaskHandle with serviceHandle.
func NewServiceHandle(handle TaskHandle) TaskHandle {
	return &serviceHandle{
		TaskHandle: handle,
		mutex:      &sync.Mutex{}}
}

// Stop implements TaskHandle interface.
func (s *serviceHandle) Stop() error {
	err := s.checkError()
	if err != nil {
		return err
	}

	return s.TaskHandle.Stop()
}

// Wait implements TaskHandle interface.
func (s *serviceHandle) Wait(duration time.Duration) bool {
	err := s.checkError()
	if err != nil {
		// TODO(skonefal): Return error here when wait will return errors.
		return s.TaskHandle.Wait(duration)
	}

	return s.TaskHandle.Wait(duration)
}

func (s *serviceHandle) checkError() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.err != nil {
		return s.err
	}

	if !s.taskCanBeTerminated {
		s.taskCanBeTerminated = true
		if s.TaskHandle.Status() == TERMINATED {
			err := errors.Errorf("ServiceHandle with command %q has terminated prematurely", s.TaskHandle.Name())
			s.err = err

			logOutput(s.TaskHandle)
			logrus.Errorf(err.Error())
			return err
		}
	}
	return nil
}
