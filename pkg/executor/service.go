package executor

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/pkg/errors"
)

var (
	// ErrServiceStopped indicates that task supposed to run endlessly stopped unexpectedly.
	ErrServiceStopped = errors.New("Task is not running")

	// LogLinesCount is the number of lines printed from stderr & stdout in case of task failure.
	LogLinesCount = conf.NewIntFlag("output_lines_count", "Number of lines printed from stderr & stdout in case of task unsucessful termination", 5)
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

// Name implements Launcher interface.
func (sl ServiceLauncher) Name() string {
	return fmt.Sprintf("Service: %q", sl.Launcher.Name())
}

// ServiceHandle is a decorator and TaskHandle implementation that should be used with tasks that do not stop on their own.
type ServiceHandle struct {
	TaskHandle
}

// Stop implements TaskHandle interface.
func (s ServiceHandle) Stop() error {
	if s.TaskHandle.Status() != RUNNING {
		logrus.Errorf("Stop(): ServiceHandle terminated prematurely")
		logOutput(s.TaskHandle)
		return ErrServiceStopped
	}

	return s.TaskHandle.Stop()
}

// Wait implements TaskHandle interface.
func (s ServiceHandle) Wait(duration time.Duration) bool {
	if s.TaskHandle.Status() != RUNNING {
		logrus.Errorf("Wait(): ServiceHandle terminated prematurely")
		logOutput(s.TaskHandle)
	}

	return s.TaskHandle.Wait(duration)
}
