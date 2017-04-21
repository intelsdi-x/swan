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
	"fmt"
	"os/exec"

	"bytes"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// Rdtset is an instance of Decorator that used rdtset command for isolation. It allows to set CPU affinity and allocate cache available to those CPUs.
// See documentation at: experiments/memcached-cat/README.md
type Rdtset struct {
	CPURange string
	Mask     int
}

// Decorate implements Decorator interface
func (r Rdtset) Decorate(command string) (decorated string) {
	decorated = fmt.Sprintf("rdtset -v -c %s -t 'l3=%#x;cpu=%s' %s", r.CPURange, r.Mask, r.CPURange, command)
	logrus.Debugf("Command decorated with rdtset: %s", decorated)

	return
}

// CleanRDTAssingments cleans any existing RDT RMID's assignment.
func CleanRDTAssingments() (string, error) {
	cmd := exec.Command("pqos", "-R")
	outputtedBytes, err := cmd.CombinedOutput()
	buf := bytes.NewBuffer(outputtedBytes)
	output := buf.String()
	if err != nil {
		return output, errors.Wrapf(err, "pqos -R failed. Output: %q", output)
	}

	return output, nil
}
