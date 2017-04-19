package executor

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

/**
ServiceLauncher and ServiceHandle are wrappers that could be used on Launcher and TaskHandle class.
User should use them to state intent that these processes should not stop without
explicit `Stop()` or `Wait()` invoked on TaskHandle.

If process would stop on it's own, the Stop() and Wait() functions will return error
and process logs will be available on experiment log stream.
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

	return &ServiceHandle{th}, nil
}

// ServiceHandle is a decorator and TaskHandle implementation that should be used with tasks that do not stop on their own.
type ServiceHandle struct {
	TaskHandle
}

// Stop implements TaskHandle interface.
func (s ServiceHandle) Stop() error {
	if s.TaskHandle.Status() != RUNNING {
		logrus.Errorf("Stop(): ServiceHandle with command %q has terminated prematurely", s.TaskHandle.Name())
		logOutput(s.TaskHandle)
		return errors.Errorf("Service %q has ended prematurely", s.TaskHandle.Name())
	}

	return s.TaskHandle.Stop()
}

// Wait implements TaskHandle interface.
func (s ServiceHandle) Wait(duration time.Duration) bool {
	if s.TaskHandle.Status() != RUNNING {
		logrus.Errorf("Wait(): ServiceHandle with command %q has terminated prematurely", s.TaskHandle.Name())
		logOutput(s.TaskHandle)
	}

	return s.TaskHandle.Wait(duration)
}
