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

package testhelpers

import (
	"fmt"
	"math/rand"
	"os/exec"
	"time"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
)

// Snapteld represents Snap daemon used in tests.
type Snapteld struct {
	task    executor.TaskHandle
	apiPort int
	rpcPort int
}

// NewSnapteld constructs Snapteld on random high ports.
func NewSnapteld() *Snapteld {
	randomHighAPIPort := rand.Intn(32768-10000) + 10000
	randomHighRPCPort := rand.Intn(42768-32768) + 32768
	return NewSnapteldOnPort(randomHighAPIPort, randomHighRPCPort)
}

// NewSnapteldOnPort constructs Snapteld on chosen ports.
func NewSnapteldOnPort(apiPort, rpcPort int) *Snapteld {
	return &Snapteld{apiPort: apiPort, rpcPort: rpcPort}
}

// NewSnapteldOnDefaultPorts constructs Snapteld on chosen ports.
func NewSnapteldOnDefaultPorts() *Snapteld {
	// Snapteld default ports are 8181(API port) and 8082(RPC port).
	return NewSnapteldOnPort(8181, 8082)
}

// Start starts Snap daemon and wait until it is responsive.
func (s *Snapteld) Start() error {
	l := executor.NewLocal()

	snapteldPath, err := exec.LookPath("snapteld")
	if err != nil {
		return errors.New("cannot find snapteld in $PATH")
	}

	snapCommand := fmt.Sprintf("%s --plugin-trust 0 --api-port %d --control-listen-port %d --log-level 1", snapteldPath, s.apiPort, s.rpcPort)

	taskHandle, err := l.Execute(snapCommand)
	if err != nil {
		return err
	}

	if !s.Connected() {
		taskHandle.Stop()
		taskHandle.EraseOutput()
		return errors.Errorf("could not connect to snapteld on %q", s.getSnapteldAddress())
	}

	s.task = taskHandle
	return nil
}

// Stop stops Snap daemon.
func (s *Snapteld) Stop() error {
	if s.task == nil {
		return errors.New("Snapteld not started: cannot find task")
	}

	return s.task.Stop()
}

// CleanAndEraseOutput cleans and removes Output.
func (s *Snapteld) CleanAndEraseOutput() error {
	if s.task == nil {
		return errors.New("Snapteld not started: cannot find task")
	}

	return s.task.EraseOutput()
}

// Connected checks if we can connect to Snap daemon.
func (s *Snapteld) Connected() bool {
	retries := 100
	isConnected := false
	cli, err := client.New(s.getSnapteldAddress(), "v1", true)
	if err != nil {
		return isConnected
	}
	for i := 0; i < retries; i++ {
		if cli.GetPlugins(false).Err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		isConnected = true
	}

	return isConnected
}

// Port returns port Snapteld is listening.
func (s *Snapteld) Port() int {
	return s.apiPort
}

func (s *Snapteld) getSnapteldAddress() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.apiPort)
}
