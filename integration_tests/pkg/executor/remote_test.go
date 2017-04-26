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
	"os/exec"
	"strings"
	"testing"

	. "github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/random"
	. "github.com/smartystreets/goconvey/convey"
)

// This tests required following setup:
// - id_rsa ssh keys in user home directory. [command ssh-keygen]
// - no password ssh session. [command ssh-copy-id localhost]
func TestRemote(t *testing.T) {
	Convey("Preparing Remote Executor to be tested on localhost", t, func() {

		remote, err := NewRemoteFromIP("127.0.0.1")
		if err != nil {
			t.Skip("Skipping remote executor test: " + err.Error())
		}

		Convey("And while using Remote Shell, the generic Executor test should pass", func() {
			testExecutor(t, remote)
		})
	})
}

func TestRemoteProcessPidIsolation(t *testing.T) {
	const memcachedBinary = "memcached"
	const mutilateBinary = "mutilate"

	// Skip test when binaries are not available.
	_, err := exec.LookPath(memcachedBinary)
	if err != nil {
		t.Skip("Skipping remote executor test: " + err.Error())
	}

	_, err = exec.LookPath(mutilateBinary)
	if err != nil {
		t.Skip("Skipping remote executor test: " + err.Error())
	}

	Convey("I should be able to execute remote command and see the processes running", t, func() {
		config := DefaultRemoteConfig()
		remote, err := NewRemote("127.0.0.1", config)
		if err != nil {
			t.Skip("Skipping remote executor test: " + err.Error())
		}

		user := config.User
		ports := random.Ports(1)
		handle, err := remote.Execute(fmt.Sprintf("%s -u %s -p %d -d && %s -A",
			memcachedBinary, user, ports[0], mutilateBinary))
		So(err, ShouldBeNil)

		mcProcCount := findProcessCount(memcachedBinary)
		So(mcProcCount, ShouldEqual, 1)
		mutProcCount := findProcessCount(mutilateBinary)
		So(mutProcCount, ShouldEqual, 1)

		Convey("I should be able to stop remote task and all the processes should be terminated", func() {
			err = handle.Stop()
			So(err, ShouldBeNil)

			mcProcCountAfterStop := findProcessCount(memcachedBinary)
			So(mcProcCountAfterStop, ShouldEqual, 0)
			mutProcCountAfterStop := findProcessCount(mutilateBinary)
			So(mutProcCountAfterStop, ShouldEqual, 0)
		})
	})
}

func findProcessCount(processName string) int {
	const separator = ","

	cmd := exec.Command("pgrep", processName, "-d "+separator)
	output, _ := cmd.Output()

	if len(output) == 0 {
		return 0
	}

	pids := strings.Split(string(output), ",")
	return len(pids)
}
