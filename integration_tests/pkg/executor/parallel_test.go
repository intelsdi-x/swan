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
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParallel(t *testing.T) {
	file, err := ioutil.TempFile(".", "parallel")
	if err != nil {
		t.Fail()
	}
	defer os.Remove(file.Name())

	Convey("When using Parallel to decorate local executor", t, func() {
		parallel := executor.NewLocalIsolated(executor.NewParallel(5))
		Convey("Process should be executed 5 times", func() {
			cmdStr := fmt.Sprintf("tailf %s", file.Name())
			task, err := parallel.Execute(cmdStr)
			defer task.EraseOutput()
			defer task.Stop()

			So(err, ShouldBeNil)
			So(task, ShouldNotBeNil)
			// NOTE: We have to wait a bit for parallel to launch commands, though.
			isStopped, err := task.Wait(1000 * time.Millisecond)
			So(err, ShouldBeNil)
			So(isStopped, ShouldBeFalse)

			cmd := exec.Command("pgrep", "-f", cmdStr)
			output, err := cmd.CombinedOutput()
			So(err, ShouldBeNil)

			pids := strings.Split(strings.TrimSpace(string(output)), "\n")
			So(len(pids), ShouldBeGreaterThan, 0)
			Convey("When I stop parallel process", func() {
				err = task.Stop()

				So(err, ShouldBeNil)
				Convey("All the child processes should be stopped", func() {
					isStopped, err := task.Wait(0)
					So(err, ShouldBeNil)
					So(isStopped, ShouldBeTrue)
					cmd = exec.Command("pgrep", "-f", cmdStr)
					err = cmd.Run()

					So(err, ShouldNotBeNil)
					So(cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus(), ShouldEqual, 1)
				})
			})
		})
	})
}
