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

package isolation

import (
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"

	"github.com/pkg/errors"
)

// CPUShares defines data needed for CPU controller.
type CPUShares struct {
	name   string
	shares int
}

// NewCPUShares instance creation.
func NewCPUShares(name string, shares int) Isolation {
	return &CPUShares{name: name, shares: shares}
}

// Decorate implements Decorator interface
func (cpu *CPUShares) Decorate(command string) string {
	return "cgexec -g cpu:" + cpu.name + " " + command
}

// Clean removes the specified cgroup
func (cpu *CPUShares) Clean() error {
	cmd := exec.Command("sh", "-c", "cgdelete -g cpu"+":"+cpu.name)
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "running command %q failed", cmd.Path)
	}

	return nil
}

// Create specified cgroup.
func (cpu *CPUShares) Create() error {
	// 1 Create cpu cgroup
	cmd := exec.Command("cgcreate", "-g", "cpu:"+cpu.name)
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "running command %q failed", cmd.Path)
	}

	// 2 Set cpu cgroup shares
	cmd = exec.Command("cgset", "-r", "cpu.shares="+strconv.Itoa(cpu.shares), cpu.name)
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "running command %q failed", cmd.Path)
	}

	return nil
}

// Isolate associates specified pid to the cgroup.
func (cpu *CPUShares) Isolate(PID int) error {
	// Associate task with the specified cgroup.
	strPID := strconv.Itoa(PID)
	d := []byte(strPID)
	filePath := path.Join("/sys/fs/cgroup/cpu", cpu.name, "tasks")
	err := ioutil.WriteFile(filePath, d, 0644)

	if err != nil {
		return errors.Wrapf(err, "could not write %q to file %q", d, filePath)
	}

	return nil
}
