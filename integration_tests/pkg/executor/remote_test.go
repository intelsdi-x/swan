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
	"os/exec"
	"strings"
	"testing"

	. "github.com/intelsdi-x/swan/pkg/executor"
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

func TestRemoteStopDetachedProcess(t *testing.T) {
	Convey("I should be able to execute remote command and see the processes running", t, func() {
		config := DefaultRemoteConfig()
		remote, err := NewRemote("127.0.0.1", config)
		if err != nil {
			t.Skip("Skipping remote executor test: " + err.Error())
		}

		sleepProcCount := findProcessCount("sleep")
		So(sleepProcCount, ShouldEqual, 0)

		handle, err := remote.Execute("sleep 1d & sleep 1d")
		So(err, ShouldBeNil)

		sleepProcCount = findProcessCount("sleep")
		So(sleepProcCount, ShouldEqual, 2)

		Convey("I should be able to stop remote task and all the processes should be terminated", func() {
			err = handle.Stop()
			So(err, ShouldBeNil)

			sleepProcCountAfterStop := findProcessCount("sleep")
			So(sleepProcCountAfterStop, ShouldEqual, 0)
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
