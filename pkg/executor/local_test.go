package executor

import (
	log "github.com/Sirupsen/logrus"
	isolation "github.com/intelsdi-x/swan/pkg/isolation"
	. "github.com/smartystreets/goconvey/convey"
	"os/user"
	"testing"
)

// TestLocal tests the execution of process on local machine.
func TestLocal(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell", t, func() {
		l := NewLocal()

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

		cpuset1 := isolation.NewCPUSet("/A", isolation.NewSet(0), isolation.NewSet(0))
		cpuset1.Create()

		l := NewIsolatedLocal([]isolation.Isolation{cpuset1})
		task, err := l.Execute("/bin/echo foobar")
		So(err, ShouldBeNil)
		task.Wait(0)
		taskState := task.Status()
		So(taskState, ShouldEqual, TERMINATED)
		exitcode, err := task.ExitCode()
		So(err, ShouldBeNil)
		So(exitcode, ShouldEqual, 0)
	})
}
