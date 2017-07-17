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
	"os/user"
	"testing"

	. "github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/cgroup"
	. "github.com/smartystreets/goconvey/convey"
)

// TestLocal tests the execution of process on local machine.
func TestLocal(t *testing.T) {
	Convey("While using Local Shell", t, func() {

		l := NewLocal()

		Convey("The generic Executor test should pass", func() {
			testExecutor(t, l)
		})
	})

	Convey("Local Executor with decorations", t, func() {
		Convey("Should run properly when no decorations are used", func() {
			l := NewLocalIsolated()
			_, err := l.Execute("echo NewLocalIsolated")
			So(err, ShouldBeNil)
		})

		Convey("Should run properly when no single decoration is used", func() {
			taskSet := isolation.Taskset{CPUList: isolation.NewIntSet(1, 2)}
			l := NewLocalIsolated(taskSet)
			_, err := l.Execute("echo NewLocalIsolated")
			So(err, ShouldBeNil)
		})

		Convey("Should run properly when multiple decorations are used", func() {
			taskSet := isolation.Taskset{CPUList: isolation.NewIntSet(1, 2)}
			l := NewLocalIsolated(taskSet, taskSet)
			_, err := l.Execute("echo NewLocalIsolated")
			So(err, ShouldBeNil)
		})
	})

	Convey("While using Local Shell using cgroups", t, func() {
		user, err := user.Current()
		if err != nil {
			t.Fatalf("Cannot get current user")
		}

		if user.Name != "root" {
			t.Skipf("Need to be privileged user to run cgroups tests")
		}

		cmd := exec.Command("cgexec")
		err = cmd.Run()
		if err != nil {
			t.Skipf("%s", err)
		}

		Convey("Creating a single cgroup with cpu set for core 0 numa node 0", func() {
			cpuset, err := cgroup.NewCPUSet("/A", isolation.NewIntSet(0), isolation.NewIntSet(0), false, false)
			So(err, ShouldBeNil)
			cpuset.Create()
			defer cpuset.Clean()

			l := NewLocalIsolated(cpuset)
			task, err := l.Execute("/bin/echo foobar")
			So(err, ShouldBeNil)
			defer task.EraseOutput()

			// Wait until command has terminated.
			task.Wait(0)

			// Ensure task is not running any longer.
			taskState := task.Status()
			So(taskState, ShouldEqual, TERMINATED)

			// Verify that the exit code represents successful run (exit code 0).
			exitcode, err := task.ExitCode()
			So(err, ShouldBeNil)
			So(exitcode, ShouldEqual, 0)
		})

		Convey("Creating a two cgroups (cpu shares and memory) for one command", func() {
			shares := isolation.NewCPUShares("/A", 1024)
			shares.Create()
			defer shares.Clean()

			memory := isolation.NewMemorySize("/A", 64*1024*1024)
			memory.Create()
			defer memory.Clean()

			l := NewLocalIsolated(isolation.Decorators{shares, memory})
			task, err := l.Execute("/bin/echo foobar")
			So(err, ShouldBeNil)
			defer task.EraseOutput()

			// Wait until command has terminated.
			task.Wait(0)

			// Ensure task is not running any longer.
			taskState := task.Status()
			So(taskState, ShouldEqual, TERMINATED)

			// Verify that the exit code represents successful run (exit code 0).
			exitcode, err := task.ExitCode()
			So(err, ShouldBeNil)
			So(exitcode, ShouldEqual, 0)
		})

		Convey("Creating nested cgroups with cpu shares", func() {
			sharesA := isolation.NewCPUShares("/A", 1024)
			sharesA.Create()
			defer sharesA.Clean()

			sharesB := isolation.NewCPUShares("/A/B", 1024)
			sharesB.Create()
			defer sharesB.Clean()

			sharesC := isolation.NewCPUShares("/A/C", 1024)
			sharesC.Create()
			defer sharesC.Clean()

			// First command.
			l1 := NewLocalIsolated(sharesB)
			task1, err := l1.Execute("/bin/echo foobar")
			So(err, ShouldBeNil)
			defer task1.EraseOutput()

			// Wait until command has terminated.
			task1.Wait(0)

			// Ensure task is not running any longer.
			taskState1 := task1.Status()
			So(taskState1, ShouldEqual, TERMINATED)

			// Verify that the exit code represents successful run (exit code 0).
			exitcode1, err := task1.ExitCode()
			So(err, ShouldBeNil)
			So(exitcode1, ShouldEqual, 0)

			// Second command.
			l2 := NewLocalIsolated(sharesC)
			task2, err := l2.Execute("/bin/echo foobar")
			So(err, ShouldBeNil)
			defer task2.EraseOutput()

			// Wait until command has terminated.
			task2.Wait(0)

			// Ensure task is not running any longer.
			taskState2 := task2.Status()
			So(taskState2, ShouldEqual, TERMINATED)

			// Verify that the exit code represents successful run (exit code 0).
			exitcode2, err := task2.ExitCode()
			So(err, ShouldBeNil)
			So(exitcode2, ShouldEqual, 0)
		})
	})
}
