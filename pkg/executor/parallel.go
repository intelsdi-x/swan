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
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/isolation"
)

// Parallel allows to run same command using same executor multiple times.
// Using Parallel decorator will mix output from all the commands executed.
// Parallel is run in new PID namespace (using isolation.Namespace) as children might not be killed correctly otherwise.
type Parallel struct {
	numberOfClones int
}

// NewParallel prepares instance of Decorator that allows to ran tasks in parallel.
func NewParallel(numberOfClones int) Parallel {
	return Parallel{numberOfClones: numberOfClones}
}

// Decorate implements isolation.Decorator interface by adding invocation of parallel to a command.
func (p Parallel) Decorate(command string) string {
	logrus.Debugf("Attempting to run command %q %d times", command, p.numberOfClones)
	decorated := fmt.Sprintf("seq %d | xargs -P%d -l -i sh -c %q", p.numberOfClones, p.numberOfClones, command)
	// You need to run parallel in new PID namespace to make sure that all the children are killed.
	unshare, err := isolation.NewNamespace(syscall.CLONE_NEWPID)
	if err != nil {
		logrus.Errorf("Impossible to create namespace decorator: %q", err)
		return command
	}
	decorated = unshare.Decorate(decorated)
	logrus.Debugf("Parallelized command prepared: %q", decorated)
	logrus.Debug("Be aware that using Parallel decorator will mix output from all the commands executed")

	return decorated
}
