package provisioning

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
	"github.com/intelsdi-x/swan/pkg/isolation"
)

func addNewline(text string) string {
	// Stdout & Stderr have a newline at the end.
	return text + `
`
}


// TestLocal takes fixed amount of time (6s) since it tests command execution and
// wait functionality.
func TestLocal(t *testing.T) {
	Convey("Using Local Shell with no isolation", t, func() {
		l := NewLocal([]isolation.ProcessIsolation{})

		Convey("When command `sleep 1` is executed and we wait for it", func() {
			start := time.Now()

			task, err := l.Run("sleep 1")

			taskNotTimeouted := task.Wait(3000)

			duration := time.Since(start)
			durationsMs := duration.Nanoseconds() / 1e6

			Convey("The command Duration should last longer than 1s", func() {
				So(durationsMs, ShouldBeGreaterThan, 1000)
			})

			Convey("And the exit status should be zero", func() {
				So(task.Status().code, ShouldEqual, 0)
			})

			Convey("And the timeout should NOT exceed", func() {
				So(taskNotTimeouted, ShouldBeTrue)
			})

			Convey("And error is nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When command `sleep 1` is executed and we wait for it with timeout 0.5s", func() {
			start := time.Now()

			task, err := l.Run("sleep 1")

			taskNotTimeouted := task.Wait(500)

			duration := time.Since(start)
			durationsMs := duration.Nanoseconds() / 1e6

			Convey("The Duration should last less than 1s", func() {
				So(durationsMs, ShouldBeLessThan, 1000)
			})

			Convey("And the exit status should point that task is still running", func() {
				So(task.Status().code, ShouldEqual, RunningCode)
			})

			Convey("And the timeout should exceed", func() {
				So(taskNotTimeouted, ShouldBeFalse)
			})

			Convey("And error is nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When command `sleep 1` is executed and we stop it after start", func() {
			start := time.Now()

			task, err := l.Run("sleep 1")

			task.Stop()

			duration := time.Since(start)
			durationsMs := duration.Nanoseconds() / 1e6

			Convey("The Duration should last less than 1s", func() {
				So(durationsMs, ShouldBeLessThan, 1000)
			})

			Convey("And the exit status should be -1", func() {
				So(task.Status().code, ShouldEqual, -1)
			})

			Convey("And error is nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When command `echo output` is executed and we wait for it", func() {
			task, err := l.Run("echo output")

			taskNotTimeouted := task.Wait(500)

			Convey("The command stdout needs to match 'output", func() {
				So(task.Status().stdout, ShouldEqual, addNewline("output"))
			})

			Convey("And the exit status should be zero", func() {
				So(task.Status().code, ShouldEqual, 0)
			})

			Convey("And the timeout should NOT exceed", func() {
				So(taskNotTimeouted, ShouldBeTrue)
			})

			Convey("And error is nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When command which does not exists is executed and we wait for it", func() {
			task, err := l.Run("commandThatDoesNotExists")

			taskNotTimeouted := task.Wait(500)

			Convey("The command stderr should point that the command does not exists", func() {
				So(task.Status().stderr, ShouldEqual,
				   addNewline("sh: 1: commandThatDoesNotExists: not found"))
			})

			Convey("The exit status should be 127", func() {
				So(task.Status().code, ShouldEqual, 127)
			})

			Convey("And the timeout should NOT exceed", func() {
				So(taskNotTimeouted, ShouldBeTrue)
			})

			Convey("And error is nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When we run two tasks in the same time", func() {
			task, err := l.Run("echo output1")
			task2, err2 := l.Run("echo output2")

			task.Wait(0)
			task2.Wait(0)

			Convey("The commands stdouts needs to match 'output1' & 'output2'", func() {
				So(task.Status().stdout, ShouldEqual, addNewline("output1"))
				So(task2.Status().stdout, ShouldEqual, addNewline("output2"))
			})

			Convey("Both exit statuses should be 0", func() {
				So(task.Status().code, ShouldEqual, 0)
				So(task2.Status().code, ShouldEqual, 0)
			})

			Convey("And errors are nil", func() {
				So(err, ShouldBeNil)
				So(err2, ShouldBeNil)
			})
		})
	})
}
