package executor

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
)

func getLocalStdoutPath(task *localTask) (string, error) {
	if _, err := os.Stat(task.stdoutFile.Name()); err != nil {
		return "", err
	}
	return task.stderrFile.Name(), nil
}

func getLocalStderrPath(task *localTask) (string, error) {
	if _, err := os.Stat(task.stderrFile.Name()); err != nil {
		return "", err
	}
	return task.stderrFile.Name(), nil
}

// TestLocal tests the execution of process on local machine.
func TestLocal(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell", t, func() {
		l := NewLocal()

		Convey("When blocking infinitively sleep command is executed", func() {
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

			Convey("When we wait for task termination with the 1ms timeout", func() {
				isTaskTerminated := task.Wait(1)

				Convey("The timeout should exceed and the task not terminated ", func() {
					So(isTaskTerminated, ShouldBeFalse)
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

				Convey("The task should be terminated and the task status should be -15", func() {
					taskState, taskStatus := task.Status()
					So(taskState, ShouldEqual, TERMINATED)
					So(*taskStatus, ShouldEqual, -15)
				})
			})
		})

		Convey("When command `echo output` is executed", func() {
			task, err := l.Execute("echo output")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				task.Stop()
			})

			Convey("When we wait for the task to terminate", func() {
				isTaskTerminated := task.Wait(500)

				Convey("Wait should states that task terminated", func() {
					So(isTaskTerminated, ShouldBeTrue)
				})

				taskState, taskStatus := task.Status()

				Convey("The task should be terminated", func() {
					So(taskState, ShouldEqual, TERMINATED)
				})

				Convey("And the exit status should be 0", func() {
					So(*taskStatus, ShouldEqual, 0)
				})

				stdoutReader := task.Stdout()
				data, err := ioutil.ReadAll(stdoutReader)
				So(err, ShouldBeNil)
				Convey("And command stdout needs to match 'output", func() {
					So(string(data[:]), ShouldEqual, "output\n")
				})
				fileName, err := getLocalStdoutPath(task.(*localTask))
				pwd, err := os.Getwd()
				So(err, ShouldBeNil)
				fileInf, err := os.Stat(fileName)
				Convey("before cleaning file should exist", func() {
					So((pwd + "/" + fileInf.Name()), ShouldEqual, fileName)
				})
				task.Clean()
				_, err = os.Stat(fileName)
				Convey("after cleaning file should not exist", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("When command which does not exists is executed", func() {
			task, err := l.Execute("commandThatDoesNotExists")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				task.Stop()
			})

			Convey("When we wait for the task to terminate", func() {
				isTaskTerminated := task.Wait(500)

				Convey("Wait should state that task terminated", func() {
					So(isTaskTerminated, ShouldBeTrue)
				})

				taskState, taskStatus := task.Status()

				Convey("The task should be terminated", func() {
					So(taskState, ShouldEqual, TERMINATED)
				})

				Convey("And the exit status should be 127", func() {
					So(*taskStatus, ShouldEqual, 127)
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

			Convey("When we wait for the tasks to terminate", func() {
				isTaskTerminated := task.Wait(0)
				isTaskTerminated2 := task2.Wait(0)

				Convey("Wait should state that tasks are terminated", func() {
					So(isTaskTerminated, ShouldBeTrue)
					So(isTaskTerminated2, ShouldBeTrue)
				})

				taskState1, taskStatus1 := task.Status()
				taskState2, taskStatus2 := task2.Status()

				Convey("The tasks should be terminated", func() {
					So(taskState1, ShouldEqual, TERMINATED)
					So(taskState2, ShouldEqual, TERMINATED)
				})

				stdoutReader := task.Stdout()
				data, err := ioutil.ReadAll(stdoutReader)
				So(err, ShouldBeNil)
				Convey("The first command stdout needs to match 'output1", func() {
					So(string(data[:]), ShouldEqual, "output1\n")
				})
				stdoutReader = task2.Stdout()
				data, err = ioutil.ReadAll(stdoutReader)
				So(err, ShouldBeNil)
				Convey("The second command stdout needs to match 'output2", func() {
					So(string(data[:]), ShouldEqual, "output2\n")
				})

				Convey("Both exit statuses should be 0", func() {
					So(*taskStatus1, ShouldEqual, 0)
					So(*taskStatus2, ShouldEqual, 0)
				})
			})
		})
	})
}
