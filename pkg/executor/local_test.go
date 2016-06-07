package executor_test // avoids import cycle when importing from cgroup package

import (
	"fmt"
	"os/exec"
	"os/user"
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/intelsdi-x/swan/pkg/executor"
	isolation "github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/cgroup"
	"github.com/pivotal-golang/bytefmt"
	. "github.com/smartystreets/goconvey/convey"
)

// TestLocal tests the execution of process on local machine.
func TestLocal(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell", t, func() {

		l := NewLocal()
		fmt.Printf("\n%q\n", l)

		Convey("The generic Executor test should pass", func() {
			testExecutor(t, l)
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

			memory := isolation.NewMemorySize("/A", 64*bytefmt.MEGABYTE)
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
