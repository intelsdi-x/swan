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

// MemorySize defines input data
type MemorySize struct {
	name string
	size int
}

// NewMemorySize creates an instance of input data.
func NewMemorySize(name string, size int) Isolation {
	return &MemorySize{
		name: name,
		size: size,
	}
}

// Decorate implements Decorator interface.
func (memorySize *MemorySize) Decorate(command string) string {
	return "cgexec -g memory:" + memorySize.name + " " + command
}

// Clean removes specified cgroup.
func (memorySize *MemorySize) Clean() error {
	cmd := exec.Command("cgdelete", "-g", "memory:"+memorySize.name)
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "running command %q failed", cmd.Path)
	}

	return nil
}

// Create specified cgroup.
func (memorySize *MemorySize) Create() error {
	// 1.a Create memory size cgroup.
	cmd := exec.Command("cgcreate", "-g", "memory:"+memorySize.name)
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "running command %q failed", cmd.Path)
	}

	// 1.b Set cgroup memory size.
	cmd = exec.Command("cgset", "-r", "memory.limit_in_bytes="+strconv.Itoa(memorySize.size), memorySize.name)
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "running command %q failed", cmd.Path)
	}

	return nil
}

// Isolate create specified cgroup and associates specified process id
func (memorySize *MemorySize) Isolate(PID int) error {
	// Set PID to cgroups.
	strPID := strconv.Itoa(PID)
	d := []byte(strPID)

	filePath := path.Join("/sys/fs/cgroup/memory", memorySize.name, "tasks")
	err := ioutil.WriteFile(filePath, d, 0644)

	if err != nil {
		return errors.Wrapf(err, "could not write %q to file %q", d, filePath)
	}

	return nil
}
