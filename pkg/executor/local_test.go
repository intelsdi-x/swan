package executor

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

// TestLocal tests the execution of process on local machine.
func TestLocal(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	Convey("While using Local Shell", t, func() {
		l := NewLocal()

		Convey("When blocking infinitively sleep command is executed", func() {
			task, err := l.Execute("sleep inf")

			Convey("There should be no error", func() {
				stopErr := task.Stop()

				So(err, ShouldBeNil)
				So(stopErr, ShouldBeNil)
			})

			Convey("Task should be still running and status should be nil", func() {
				taskState, taskStatus := task.Status()
				So(taskState, ShouldEqual, RUNNING)
				So(taskStatus, ShouldBeNil)

				stopErr := task.Stop()
				So(stopErr, ShouldBeNil)

			})

			Convey("When we wait for task termination with the 1ms timeout", func() {
				isTaskTerminated := WaitWithTimeout(task, 1*time.Microsecond)

				Convey("The timeout should exceed and the task not terminated ", func() {
					So(isTaskTerminated, ShouldBeFalse)
				})

				Convey("The task should be still running and status should be nil", func() {
					taskState, taskStatus := task.Status()
					So(taskState, ShouldEqual, RUNNING)
					So(taskStatus, ShouldBeNil)
				})

				stopErr := task.Stop()
				So(stopErr, ShouldBeNil)
			})

			Convey("When we stop the task", func() {
				err := task.Stop()

				Convey("There should be no error", func() {
					So(err, ShouldBeNil)
				})

				Convey("The task should be terminated and the task status should be 143", func() {
					taskState, taskStatus := task.Status()

					So(taskState, ShouldEqual, TERMINATED)
					So(taskStatus, ShouldNotBeNil)
					So(taskStatus.ExitCode, ShouldEqual, 143)
				})
			})

			Convey("When multpile go routines waits for task termination", func() {
				waitErrChannel := make(chan error)
				for i := 0; i < 5; i++ {
					go func() {
						waitErrChannel <- task.Wait()
					}()
				}

				Convey("All waits should be blocked", func() {
					gotWaitResult := false
					select {
					case _ = <-waitErrChannel:
						gotWaitResult = true
					default:
					}

					err := task.Stop()
					So(err, ShouldBeNil)
					So(gotWaitResult, ShouldBeFalse)

					Convey("After stop every wait in go routine should result with nil", func() {
						for i := 0; i < 5; i++ {
							err = <-waitErrChannel
							So(err, ShouldBeNil)
						}
					})
				})
			})
		})

		Convey("When command `echo output` is executed", func() {
			task, err := l.Execute("echo output")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				err = task.Stop()
				So(err, ShouldBeNil)
			})

			Convey("When we wait for the task to terminate", func() {
				task.Wait()

				taskState, taskStatus := task.Status()

				Convey("The task should be terminated", func() {
					So(taskState, ShouldEqual, TERMINATED)
				})

				Convey("And the exit status should be 0 and command needs to be 'output'", func() {
					So(taskStatus, ShouldNotBeNil)
					So(taskStatus.ExitCode, ShouldEqual, 0)
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

			Convey("When we wait for the task to terminate", func() {
				task.Wait()

				taskState, taskStatus := task.Status()

				Convey("The task should be terminated", func() {
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

			Convey("When we wait for the tasks to terminate", func() {
				task.Wait()
				task2.Wait()

				taskState1, taskStatus1 := task.Status()
				taskState2, taskStatus2 := task2.Status()

				Convey("The tasks should be terminated", func() {
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
