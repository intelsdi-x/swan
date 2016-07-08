package executor

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
)

// ClusterTaskHandle is a task handle for composition of master and multiple agents.
// - StdoutFile, StderrFile, ExitCode, Status, Address are taken from master task handle.
// - Clean, EraseOutput, Stop are done for master and all agents.
// - Wait waits for master task and when it's terminated it stops all agents.
// It implements TaskHandle interface.
type ClusterTaskHandle struct {
	master TaskHandle
	agents []TaskHandle
}

// NewClusterTaskHandle returns a ClusterTaskHandle instance.
func NewClusterTaskHandle(master TaskHandle, agents []TaskHandle) *ClusterTaskHandle {
	return &ClusterTaskHandle{
		master: master,
		agents: agents,
	}
}

// AddAgent gives possibility to dynamically add agent to already handled agents.
func (m ClusterTaskHandle) AddAgent(agent TaskHandle) {
	m.agents = append(m.agents, agent)
}

// StdoutFile returns a file handle for the master's stdout file.
func (m ClusterTaskHandle) StdoutFile() (*os.File, error) {
	return m.master.StdoutFile()
}

// StderrFile returns a file handle for the master's stderr file.
func (m ClusterTaskHandle) StderrFile() (*os.File, error) {
	return m.master.StderrFile()
}

// Stop terminates the master firstly and then all the agents.
func (m ClusterTaskHandle) Stop() (err error) {
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

// Status returns the state of the master.
func (m ClusterTaskHandle) Status() TaskState {
	return m.master.Status()
}

// ExitCode returns the master exitCode. If master is not terminated it returns error.
func (m ClusterTaskHandle) ExitCode() (int, error) {
	return m.master.ExitCode()
}

// Wait does the blocking wait for the master completion in case of 0 timeout time.
// Wait is a helper for waiting with a given timeout time.
// It returns true if all tasks are terminated.
func (m ClusterTaskHandle) Wait(timeout time.Duration) bool {
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
func (m ClusterTaskHandle) Clean() (err error) {
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
func (m ClusterTaskHandle) EraseOutput() (err error) {
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
func (m ClusterTaskHandle) Address() string {
	return m.master.Address()
}
