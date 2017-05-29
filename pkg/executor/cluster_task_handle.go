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
	"fmt"
	"os"
	"time"

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

// AddAgent dynamically adds agent to already handled agents.
func (m *ClusterTaskHandle) AddAgent(agent TaskHandle) {
	m.agents = append(m.agents, agent)
}

// StdoutFile returns a file handle for the master's stdout file.
func (m *ClusterTaskHandle) StdoutFile() (*os.File, error) {
	return m.master.StdoutFile()
}

// StderrFile returns a file handle for the master's stderr file.
func (m *ClusterTaskHandle) StderrFile() (*os.File, error) {
	return m.master.StderrFile()
}

// Stop terminates the master firstly and then all the agents.
func (m *ClusterTaskHandle) Stop() (err error) {
	var errCollection errcollection.ErrorCollection

	// Stop master.
	err = m.master.Stop()
	errCollection.Add(err)

	// Stop agents.
	for _, handle := range m.agents {
		err = handle.Stop()
		errCollection.Add(err)
	}

	return errCollection.GetErrIfAny()
}

// Status returns the state of the master.
func (m *ClusterTaskHandle) Status() TaskState {
	return m.master.Status()
}

// ExitCode returns the master exitCode. If master is not terminated it returns error.
func (m *ClusterTaskHandle) ExitCode() (int, error) {
	return m.master.ExitCode()
}

// Wait does the blocking wait for the master completion in case of 0 timeout time.
// Wait is a helper for waiting with a given timeout time.
// It returns true if all tasks are terminated.
func (m *ClusterTaskHandle) Wait(timeout time.Duration) (isMasterTerminated bool, err error) {
	// Wait for master first.
	isMasterTerminated, err = m.master.Wait(timeout)
	if err != nil {
		agentErrors := m.stopAgents()
		// We don't want to lose master error stack trace for no reason.
		if agentErrors != nil {
			var errCol errcollection.ErrorCollection
			errCol.Add(err)
			errCol.Add(agentErrors)
			err = errCol.GetErrIfAny()
		}
		return isMasterTerminated, err
	}

	if isMasterTerminated {
		err = m.stopAgents()
	}

	return isMasterTerminated, err
}

func (m *ClusterTaskHandle) stopAgents() error {
	var errCol errcollection.ErrorCollection
	// Stop the agents when master is terminated.
	for _, handle := range m.agents {
		err := handle.Stop()
		if err != nil {
			errCol.Add(err)
		}
	}
	return errCol.GetErrIfAny()
}

// EraseOutput removes master's and agents' stdout & stderr files.
func (m *ClusterTaskHandle) EraseOutput() (err error) {
	var errCollection errcollection.ErrorCollection

	err = m.master.EraseOutput()
	errCollection.Add(err)
	for _, handle := range m.agents {
		err = handle.EraseOutput()
		errCollection.Add(err)
	}

	return errCollection.GetErrIfAny()
}

// String returns name of underlying task.
func (m *ClusterTaskHandle) String() string {
	return fmt.Sprintf("Cluster TaskHandle containg master: %s", m.master.String())
}

// Address returns address of master task.
func (m *ClusterTaskHandle) Address() string {
	return m.master.Address()
}
