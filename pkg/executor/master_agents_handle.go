package executor

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"os"
	"time"
)

// MasterAgentsTaskHandle implements TaskHandle interface.
type MasterAgentsTaskHandle struct {
	master TaskHandle
	agents []TaskHandle
}

// NewMasterAgentsTaskHandle returns a MasterAgentsTaskHandle instance.
func NewMasterAgentsTaskHandle(master TaskHandle, agents []TaskHandle) *MasterAgentsTaskHandle {
	return &MasterAgentsTaskHandle{
		master: master,
		agents: agents,
	}
}

// StdoutFile returns a file handle for file to the master's stdout file.
func (m MasterAgentsTaskHandle) StdoutFile() (*os.File, error) {
	return m.master.StdoutFile()
}

// StderrFile returns a file handle for file to the master's stderr file.
func (m MasterAgentsTaskHandle) StderrFile() (*os.File, error) {
	return m.master.StderrFile()
}

// Stop terminates the master firstly and then all the agents.
func (m MasterAgentsTaskHandle) Stop() (err error) {
	var errCollection errcollection.ErrorCollection

	// Stop master.
	err = m.master.Stop()
	if err != nil {
		errCollection.Add(err)
	}

	// Stop agents.
	for _, handle := range m.agents {
		err = handle.Stop()
		if err != nil {
			errCollection.Add(err)
		}
	}

	return errCollection.GetErrIfAny()
}

// Status returns a state of the master.
func (m MasterAgentsTaskHandle) Status() TaskState {
	return m.master.Status()
}

// ExitCode returns a master exitCode. If master is not terminated it returns error.
func (m MasterAgentsTaskHandle) ExitCode() (int, error) {
	return m.master.ExitCode()
}

// Wait does the blocking wait for the master completion in case of nil.
// Wait is a helper for waiting with a given timeout time.
// It returns true if all tasks are terminated.
func (m MasterAgentsTaskHandle) Wait(timeout time.Duration) bool {
	// Wait for master first.
	isMasterTerminated := m.master.Wait(timeout)
	if isMasterTerminated {
		// Stop the agents when master is terminated.
		for _, handle := range m.agents {
			err := handle.Stop()
			if err != nil {
				// Just log the error if agent cannot be stopped.
				logrus.Error(err.Error())
			}
		}
	}
	return isMasterTerminated
}

// Clean cleans master and agents temporary resources.
// It also closes the tasks' stdout & stderr files.
func (m MasterAgentsTaskHandle) Clean() (err error) {
	var errCollection errcollection.ErrorCollection

	err = m.master.Clean()
	if err != nil {
		errCollection.Add(err)
	}
	for _, handle := range m.agents {
		err = handle.Clean()
		if err != nil {
			errCollection.Add(err)
		}
	}

	return errCollection.GetErrIfAny()
}

// EraseOutput removes master's and agents' stdout & stderr files.
func (m MasterAgentsTaskHandle) EraseOutput() (err error) {
	var errCollection errcollection.ErrorCollection

	err = m.master.EraseOutput()
	if err != nil {
		errCollection.Add(err)
	}
	for _, handle := range m.agents {
		err = handle.EraseOutput()
		if err != nil {
			errCollection.Add(err)
		}
	}

	return errCollection.GetErrIfAny()
}

// Address returns address of master task.
func (m MasterAgentsTaskHandle) Address() string {
	return m.master.Address()
}
