package executor

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/pkg/errors"
)

// ErrServiceStopped indicates that task supposed to run endlessly stopped unexpectedly.
var ErrServiceStopped = errors.New("Task is not running")

// LogLinesCount is the number of lines printed from stderr & stdout in case of task failure.
var LogLinesCount = conf.NewIntFlag("output_lines_count", "Number of lines printed from stderr & stdout in case of task unsucessful termination", 5)

func logOutput(th TaskHandle) error {
	lines := LogLinesCount.Value()
	file, err := th.StdoutFile()
	if err == nil {
		stdout, err := tailFile(file.Name(), lines)
		if err != nil {
			logrus.Errorf("Tailing stdout file failed: %q", err.Error())
		}
		logrus.Errorf("Last %d lines of stdout: %s", lines, stdout)
	} else {
		logrus.Errorf("Impossible to retrieve stdout file: %q", err.Error())
	}
	file, err = th.StderrFile()
	if err == nil {
		stderr, err := tailFile(file.Name(), lines)
		if err != nil {
			logrus.Errorf("Tailing stderr file failed: %q", err.Error())
		}

		logrus.Errorf("Last %d lines of stderr: %s", lines, stderr)
	} else {
		logrus.Errorf("Impossible to retrieve stderr file: %q", err.Error())
	}

	return nil

}

// Service is a decorator and TaskHandle implementation that should be used with tasks that do not stop on their own.
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

func tailFile(filePath string, lineCount int) (tail string, err error) {
	lineCountParam := fmt.Sprintf("-n %d", lineCount)
	output, err := exec.Command("tail", lineCountParam, filePath).CombinedOutput()

	if err != nil {
		return "", errors.Wrapf(err, "could not read tail of %q", filePath)
	}

	return string(output), nil
}

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
