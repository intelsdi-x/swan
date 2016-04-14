package executor

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const (
	fifoTestDirTemplate = "/tmp/swan_local_test.XXXXXXXXXXX"
	fifoTestName        = "swan_fifo"
)

// TestLocal
func TestLocal(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell", t, func() {
		l := NewLocal()

		Convey("When blocking infinitive sleep command "+
			"is executed", func() {
			task, err := l.Execute("sleep inf")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				task.Stop()
			})

			Convey("Task should be still running and status should be nil", func() {
				taskState, taskStatus := task.Status()
				So(taskState, ShouldEqual, RUNNING)
				So(taskStatus, ShouldBeNil)

				task.Stop()
			})

			Convey("When we wait for task with the 1ms timeout", func() {
				taskNotTimeouted := task.Wait(1)

				Convey("The timeout should exceed ", func() {
					So(taskNotTimeouted, ShouldBeFalse)
				})

				Convey("The task should be still running and status should be nil", func() {
					taskState, taskStatus := task.Status()
					So(taskState, ShouldEqual, RUNNING)
					So(taskStatus, ShouldBeNil)
				})

				task.Stop()
			})

			Convey("When we stop the task", func() {
				err := task.Stop()

				Convey("There should be no error", func() {
					So(err, ShouldBeNil)
				})

				Convey("The task should be NOT running and the task status should be -1", func() {
					taskState, taskStatus := task.Status()
					So(taskState, ShouldEqual, TERMINATED)
					So(taskStatus.ExitCode, ShouldEqual, -1)
				})
			})
		})

		Convey("When command `echo output` is executed", func() {
			task, err := l.Execute("echo output")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				task.Stop()
			})

			Convey("When we wait for the task", func() {
				taskNotTimeouted := task.Wait(500)

				Convey("The timeout should NOT exceed", func() {
					So(taskNotTimeouted, ShouldBeTrue)
				})

				taskState, taskStatus := task.Status()

				Convey("The task should be NOT running", func() {
					So(taskState, ShouldEqual, TERMINATED)
				})

				Convey("And the exit status should be 0", func() {
					So(taskStatus.ExitCode, ShouldEqual, 0)
				})

				Convey("And command stdout needs to match 'output", func() {
					So(taskStatus.Stdout, ShouldEqual, "output\n")
				})
			})
		})

		Convey("When command which does not exists is executed", func() {
			task, err := l.Execute("commandThatDoesNotExists")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				task.Stop()
			})

			Convey("When we wait for the task", func() {
				taskNotTimeouted := task.Wait(500)

				Convey("The timeout should NOT exceed", func() {
					So(taskNotTimeouted, ShouldBeTrue)
				})

				taskState, taskStatus := task.Status()

				Convey("The task should be NOT running", func() {
					So(taskState, ShouldEqual, TERMINATED)
				})

				Convey("And the exit status should be 127", func() {
					So(taskStatus.ExitCode, ShouldEqual, 127)
				})
			})
		})

		Convey("When we execute two tasks in the same time", func() {
			task, err := l.Execute("echo output1")
			task2, err2 := l.Execute("echo output2")

			Convey("There should be no errors", func() {
				So(err, ShouldBeNil)
				So(err2, ShouldBeNil)
			})

			Convey("When we wait for the tasks", func() {
				taskNotTimeouted := task.Wait(0)
				taskNotTimeouted2 := task2.Wait(0)

				Convey("The timeouts should NOT exceed", func() {
					So(taskNotTimeouted, ShouldBeTrue)
					So(taskNotTimeouted2, ShouldBeTrue)
				})

				taskState1, taskStatus1 := task.Status()
				taskState2, taskStatus2 := task2.Status()

				Convey("The tasks should be not running", func() {
					So(taskState1, ShouldEqual, TERMINATED)
					So(taskState2, ShouldEqual, TERMINATED)
				})

				Convey("The commands stdouts needs to match 'output1' & 'output2'", func() {
					So(taskStatus1.Stdout, ShouldEqual, "output1\n")
					So(taskStatus2.Stdout, ShouldEqual, "output2\n")
				})

				Convey("Both exit statuses should be 0", func() {
					So(taskStatus1.ExitCode, ShouldEqual, 0)
					So(taskStatus2.ExitCode, ShouldEqual, 0)
				})
			})
		})
	})
}
