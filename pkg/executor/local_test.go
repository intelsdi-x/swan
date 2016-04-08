package executor

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"syscall"
	"os/exec"
)

const (FifoTestPipePath = "/tmp/swan_local_test_fifo")

// TestLocal
func TestLocal(t *testing.T) {
	// Prepare fifo pipe for following tests.
	cmd := exec.Command("rm", "-rf", FifoTestPipePath)
	err := cmd.Run()
	if err != nil {
		t.Error(err)
	}

	err = syscall.Mkfifo(FifoTestPipePath, syscall.S_IFIFO)
	if err != nil {
		t.Error(err)
	}

	Convey("Using Local Shell", t, func() {
		l := NewLocal()

		Convey("When command waiting for signal in fifo is executed and we wait for it with timeout 1ms", func() {
			task, err := l.Run("read -n 1 <" + FifoTestPipePath)

			taskNotTimeouted := task.Wait(1)

			running, _ := task.Status()

			Convey("The status result should point that the task is still running", func() {
				So(running, ShouldBeTrue)
			})

			Convey("And the timeout should exceed", func() {
				So(taskNotTimeouted, ShouldBeFalse)
			})

			Convey("And error is nil", func() {
				So(err, ShouldBeNil)
			})

			task.Stop()
		})

		Convey("When command waiting for signal in fifo is executed and we stop it after start", func() {
			task, err := l.Run("read -n 1 <" + FifoTestPipePath)

			task.Stop()

			running, status := task.Status()

			Convey("The status result should point that the task is not running", func() {
				So(running, ShouldBeFalse)
			})

			Convey("And the exit status should be -1", func() {
				So(status.code, ShouldEqual, -1)
			})

			Convey("And error is nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When command `echo output` is executed and we wait for it", func() {
			task, err := l.Execute("echo output")

			taskNotTimeouted := task.Wait(500)

			running, status := task.Status()

			Convey("The status result should point that the task is not running", func() {
				So(running, ShouldBeFalse)
			})

			Convey("And the exit status should be 0", func() {
				So(status.code, ShouldEqual, 0)
			})

			Convey("And command stdout needs to match 'output", func() {
				So(status.stdout, ShouldEqual, "output\n")
			})

			Convey("And the timeout should NOT exceed", func() {
				So(taskNotTimeouted, ShouldBeTrue)
			})

			Convey("And error is nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When command which does not exists is executed and we wait for it", func() {
			task, err := l.Execute("commandThatDoesNotExists")

			taskNotTimeouted := task.Wait(500)

			running, status := task.Status()

			Convey("The status result should point that the task is not running", func() {
				So(running, ShouldBeFalse)
			})

			Convey("And the exit status should be 127", func() {
				So(status.code, ShouldEqual, 127)
			})

			Convey("And the timeout should NOT exceed", func() {
				So(taskNotTimeouted, ShouldBeTrue)
			})

			Convey("And error is nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When we execute two tasks in the same time", func() {
			task, err := l.Execute("echo output1")
			task2, err2 := l.Execute("echo output2")

			task.Wait(0)
			task2.Wait(0)

			running1, status1 := task.Status()
			running2, status2 := task2.Status()

			Convey("The status results should point that the tasks are not running", func() {
				So(running1, ShouldBeFalse)
				So(running2, ShouldBeFalse)
			})

			Convey("The commands stdouts needs to match 'output1' & 'output2'", func() {
				So(status1.stdout, ShouldEqual, "output1\n")
				So(status2.stdout, ShouldEqual, "output2\n")
			})

			Convey("Both exit statuses should be 0", func() {
				So(status1.code, ShouldEqual, 0)
				So(status2.code, ShouldEqual, 0)
			})

			Convey("And errors are nil", func() {
				So(err, ShouldBeNil)
				So(err2, ShouldBeNil)
			})
		})
	})

	// Clean up
	cmd = exec.Command("rm", "-rf", FifoTestPipePath)
	err = cmd.Run()
	if err != nil {
		t.Error(err)
	}
}
