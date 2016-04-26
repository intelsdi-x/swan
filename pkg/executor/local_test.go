package executor

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"
)

// TestLocal tests the execution of process on local machine.
func TestLocal(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

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
				So(taskStatus, ShouldEqual, -1)

				stopErr := task.Stop()
				So(stopErr, ShouldBeNil)

			})

			Convey("When we wait for task termination with the 1ms timeout", func() {
				isTaskTerminated := task.Wait(1 * time.Microsecond)

				Convey("The timeout should exceed and the task not terminated ", func() {
					So(isTaskTerminated, ShouldBeFalse)
				})

				Convey("The task should be still running and status should be nil", func() {
					taskState, taskStatus := task.Status()
					So(taskState, ShouldEqual, RUNNING)
					So(taskStatus, ShouldEqual, -1)
				})

				stopErr := task.Stop()
				So(stopErr, ShouldBeNil)
			})

			Convey("When we stop the task", func() {
				err := task.Stop()

				Convey("There should be no error", func() {
					So(err, ShouldBeNil)
				})

				Convey("The task should be terminated and the task status should be -1", func() {
					taskState, taskStatus := task.Status()

					So(taskState, ShouldEqual, TERMINATED)
					So(taskStatus, ShouldEqual, -1)
				})
			})

			Convey("When multiple go routines waits for task termination", func() {
				var wg sync.WaitGroup
				wg.Add(5)
				for i := 0; i < 5; i++ {
					go func() {
						task.Wait(0)
						wg.Done()
					}()
				}

				allWaitsAreDone := make(chan struct{})

				go func() {
					wg.Wait()
					close(allWaitsAreDone)
				}()

				Convey("All waits should be blocked", func() {
					waitsDone := false
					select {
					case <-allWaitsAreDone:
						waitsDone = true
					case <-time.After(100 * time.Millisecond):
					}

					err := task.Stop()
					So(err, ShouldBeNil)
					So(waitsDone, ShouldBeFalse)

					Convey("After stop every wait should unblock", func() {
						select {
						case <-allWaitsAreDone:
							waitsDone = true
						case <-time.After(100 * time.Millisecond):
						}

						So(waitsDone, ShouldBeTrue)
					})
				})
			})

			if task != nil {
				task.Clean()
			}
		})

		Convey("When command `echo output` is executed", func() {
			task, err := l.Execute("echo output")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				task.Stop()
			})

			Convey("When we wait for the task to terminate", func() {
				task.Wait(0)

				taskState, taskStatus := task.Status()

				Convey("The task should be terminated", func() {
					So(taskState, ShouldEqual, TERMINATED)
				})

				Convey("And the exit status should be 0 and command needs to be 'output'", func() {
					So(taskStatus, ShouldEqual, 0)

					stdoutReader, stdoutErr := task.Stdout()
					So(stdoutErr, ShouldBeNil)
					So(stdoutReader, ShouldNotBeNil)

					data, readErr := ioutil.ReadAll(stdoutReader)
					So(readErr, ShouldBeNil)
					So(string(data[:]), ShouldEqual, "output\n")

				})

				Convey("And the cleaning should clean the stdout file", func() {

					Convey("Before cleaning file should exist", func() {
						fileName := task.(*asyncTask).stdoutFile.Name()
						_, statErr := os.Stat(fileName)
						So(statErr, ShouldBeNil)
					})

					err := task.Clean()
					So(err, ShouldBeNil)

					Convey("After cleaning file should not exist", func() {
						fileName := task.(*asyncTask).stdoutFile.Name()
						_, statErr := os.Stat(fileName)
						So(statErr, ShouldNotBeNil)
					})
				})
			})

			if task != nil {
				task.Clean()
			}
		})

		Convey("When command which does not exists is executed", func() {
			task, err := l.Execute("commandThatDoesNotExists")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				task.Stop()
			})

			Convey("When we wait for the task to terminate", func() {
				task.Wait(0)

				taskState, taskStatus := task.Status()

				Convey("The task should be terminated", func() {
					So(taskState, ShouldEqual, TERMINATED)
				})

				Convey("And the exit status should be 127 and stderr mentioning not"+
					"found command", func() {
					So(taskStatus, ShouldEqual, 127)

					stderrReader, stderrErr := task.Stderr()
					So(stderrErr, ShouldBeNil)
					So(stderrReader, ShouldNotBeNil)

					data, readErr := ioutil.ReadAll(stderrReader)
					So(readErr, ShouldBeNil)
					So(string(data[:]), ShouldContainSubstring, "commandThatDoesNotExists")
				})

				Convey("And the cleaning should clean the stderr file", func() {

					Convey("Before cleaning file should exist", func() {
						fileName := task.(*asyncTask).stderrFile.Name()
						_, statErr := os.Stat(fileName)
						So(statErr, ShouldBeNil)
					})

					err := task.Clean()
					So(err, ShouldBeNil)

					Convey("After cleaning file should not exist", func() {
						fileName := task.(*asyncTask).stderrFile.Name()
						_, statErr := os.Stat(fileName)
						So(statErr, ShouldNotBeNil)
					})
				})

			})

			if task != nil {
				task.Clean()
			}
		})

		Convey("When we execute two tasks in the same time", func() {
			task, err := l.Execute("echo output1")
			task2, err2 := l.Execute("echo output2")

			Convey("There should be no errors", func() {
				So(err, ShouldBeNil)
				So(err2, ShouldBeNil)
			})

			Convey("When we wait for the tasks to terminate", func() {
				task.Wait(0)
				task2.Wait(0)

				taskState1, taskStatus1 := task.Status()
				taskState2, taskStatus2 := task2.Status()

				Convey("The tasks should be terminated", func() {
					So(taskState1, ShouldEqual, TERMINATED)
					So(taskState2, ShouldEqual, TERMINATED)
				})

				Convey("The commands stdouts needs to match 'output1' & 'output2'", func() {
					stdoutReader, stdoutErr := task.Stdout()
					So(stdoutErr, ShouldBeNil)
					So(stdoutReader, ShouldNotBeNil)

					data, readErr := ioutil.ReadAll(stdoutReader)
					So(readErr, ShouldBeNil)
					So(string(data[:]), ShouldEqual, "output1\n")

					stdoutReader, stdoutErr = task2.Stdout()
					So(stdoutErr, ShouldBeNil)
					So(stdoutReader, ShouldNotBeNil)

					data, readErr = ioutil.ReadAll(stdoutReader)
					So(readErr, ShouldBeNil)
					So(string(data[:]), ShouldEqual, "output2\n")

				})

				Convey("Both exit statuses should be 0", func() {
					So(taskStatus1, ShouldEqual, 0)
					So(taskStatus2, ShouldEqual, 0)
				})
			})

			if task != nil {
				task.Clean()
			}

			if task2 != nil {
				task2.Clean()
			}
		})

		Convey("When command `echo sleep 0` is executed", func() {
			task, err := l.Execute("echo sleep 0")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)
			})

			// Wait for the command to execute.
			time.Sleep(100 * time.Millisecond)

			Convey("When we get Status without the Wait for it", func() {
				taskState, taskStatus := task.Status()

				Convey("And the task should stated that it terminated", func() {
					So(taskState, ShouldEqual, TERMINATED)
				})

				Convey("And the exit status should be 0", func() {
					So(taskStatus, ShouldEqual, 0)
				})
			})

			if task != nil {
				task.Clean()
			}
		})
	})
}
