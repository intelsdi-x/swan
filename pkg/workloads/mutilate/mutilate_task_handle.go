package mutilate

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	"os"
	"time"
)

type MutilateTaskHandle struct {
	master executor.TaskHandle
	agents []executor.TaskHandle
}

func (m MutilateTaskHandle) StdoutFile() (*os.File, error) {
	return m.master.StdoutFile()
}

//	// StderrFile returns a file handle for file to the task's stderr file.
func (m MutilateTaskHandle) StderrFile() (*os.File, error) {
	return m.master.StderrFile()
}

//	// Stops a task.
func (m MutilateTaskHandle) Stop() (err error) {
	errorString := ""
	err = m.master.Stop()
	if err != nil {
		errorString += err.Error()
	}
	for _, handle := range m.agents {
		err = handle.Stop()
		if err != nil {
			errorString += err.Error()
		}
	}
	if errorString != "" {
		return fmt.Errorf(errorString)
	}
	return nil
}

//	// Status returns a state of the task.
func (m MutilateTaskHandle) Status() executor.TaskState {
	return m.master.Status()
}

//	// ExitCode returns a exitCode. If task is not terminated it returns error.
func (m MutilateTaskHandle) ExitCode() (int, error) {
	return m.master.ExitCode()
}

//	// Wait does the blocking wait for the task completion in case of nil.
//	// Wait is a helper for waiting with a given timeout time.
//	// It returns true if task is terminated.
func (m MutilateTaskHandle) Wait(timeout time.Duration) bool {
	isMasterTerminated := m.master.Wait(timeout)
	if isMasterTerminated {
		for _, handle := range m.agents {
			handle.Stop()
		}
	}
	return isMasterTerminated
}

//	// Clean cleans task temporary resources like isolations for Local.
//	// It also closes the task's stdout & stderr files.
func (m MutilateTaskHandle) Clean() (err error) {
	errorString := ""
	err = m.master.Clean()
	if err != nil {
		errorString += err.Error()
	}
	for _, handle := range m.agents {
		err = handle.Clean()
		if err != nil {
			errorString += err.Error()
		}
	}
	if errorString != "" {
		return fmt.Errorf(errorString)
	}
	return nil
}

//	// EraseOutput removes task's stdout & stderr files.
func (m MutilateTaskHandle) EraseOutput() (err error) {
	errorString := ""
	err = m.master.EraseOutput()
	if err != nil {
		errorString += err.Error()
	}
	for _, handle := range m.agents {
		err = handle.EraseOutput()
		if err != nil {
			errorString += err.Error()
		}
	}
	if errorString != "" {
		return fmt.Errorf(errorString)
	}
	return nil
}

//	// Location returns address where task was located.
func (m MutilateTaskHandle) Address() string {
	return m.master.Address()
}
