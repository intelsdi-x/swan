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

package cgroup

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
)

// Subsys returns true if the named subsystem is mounted.
func Subsys(name string, executor executor.Executor, timeout time.Duration) (bool, error) {
	mounts, err := SubsysMounts(executor, timeout)
	if err != nil {
		return false, err
	}
	_, found := mounts[name]
	return found, nil
}

// SubsysPath returns the absolute path where the supplied subsystem is
// mounted. Returns the empty string if the subsystem is not mounted.
func SubsysPath(name string, executor executor.Executor, timeout time.Duration) (string, error) {
	mounts, err := SubsysMounts(executor, timeout)
	if err != nil {
		return "", err
	}
	mount, mounted := mounts[name]
	if !mounted {
		return "", fmt.Errorf("Subsystem '%s' is not mounted", name)
	}
	return mount, nil
}

// SubsysMounts returns a map of cgroup subsystem controller names to
// mount points in the file system.
func SubsysMounts(executor executor.Executor, timeout time.Duration) (map[string]string, error) {
	out, err := cmdOutput(executor, timeout, "lssubsys", "--all-mount-points")
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		var name, mount string
		_, err := fmt.Sscanf(line, "%s %s", &name, &mount)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// The name part may indicate co-mounted subsystems (e.g. "cpu,cpuacct").
		// Let's save them separately to make them easier to find.
		names := strings.Split(name, ",")
		for _, n := range names {
			result[n] = mount
		}
	}
	return result, nil
}
