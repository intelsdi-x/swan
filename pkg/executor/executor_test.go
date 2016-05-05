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

// testExecutor tests the execution of process for given executor.
// This test can be used inside any Executor implementation test.
func testExecutor(t *testing.T, executor Executor) {
	log.SetLevel(log.DebugLevel)

	Convey("When blocking infinitively sleep command is executed", func() {
		task, err := executor.Execute("sleep inf")
		So(err, ShouldBeNil)

		defer task.Stop()
		defer task.Clean()
		defer task.EraseOutput()

		Convey("Task should be still running and status should be nil", func() {
			taskState, taskStatus := task.Status()
			So(taskState, ShouldEqual, RUNNING)
			So(taskStatus.Empty(), ShouldBeTrue)

			stopErr := task.Stop()
			So(stopErr, ShouldBeNil)
		})

		Convey("When we wait for task termination with the 1ms timeout", func() {
			isTaskTerminated := task.Wait(1 * time.Microsecond)

			Convey("The timeout appeach and the task should not be terminated", func() {
				So(isTaskTerminated, ShouldBeFalse)
			})

			Convey("The task should be still running and status should be nil", func() {
				taskState, taskStatus := task.Status()
				So(taskState, ShouldEqual, RUNNING)
				So(taskStatus.Empty(), ShouldBeTrue)
			})

			stopErr := task.Stop()
			So(stopErr, ShouldBeNil)
		})

		Convey("When we stop the task", func() {
			err := task.Stop()

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("The task should be terminated and the task status should "+
				"indicate that task was killed", func() {
				taskState, taskStatus := task.Status()

				So(taskState, ShouldEqual, TERMINATED)
				// -1 for Local executor.
				// 129 for Remote executor.
				// TODO: Unify exit code constants in next PR.
				So(taskStatus.Get().(int), ShouldBeIn, -1, 129)
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
	})

	Convey("When command `echo output` is executed", func() {
		task, err := executor.Execute("echo output")
		So(err, ShouldBeNil)

		defer task.Stop()
		defer task.Clean()
		defer task.EraseOutput()

		Convey("When we wait for the task to terminate", func() {
			task.Wait(0)

			taskState, taskStatus := task.Status()

			Convey("The task should be terminated", func() {
				So(taskState, ShouldEqual, TERMINATED)
			})

			Convey("And the exit status should be 0 and output needs to be 'output'", func() {
				So(taskStatus.Get().(int), ShouldEqual, 0)

				stdoutReader, stdoutErr := task.Stdout()
				So(stdoutErr, ShouldBeNil)
				So(stdoutReader, ShouldNotBeNil)

				data, readErr := ioutil.ReadAll(stdoutReader)
				So(readErr, ShouldBeNil)
				So(string(data[:]), ShouldStartWith, "output")

			})

			Convey("And the eraseOutput should clean the stdout file", func() {
				var filePath string
				switch explicitTask := task.(type) {
				case *localTask:
					filePath = explicitTask.stdoutFile.Name()
				case *remoteTask:
					filePath = explicitTask.stdoutFile.Name()
				default:
					t.Skip("Skipping test: task is neither instance of localTask nor remoteTask")

				}

				task.Clean()
				Convey("Before eraseOutput file should exist", func() {
					_, statErr := os.Stat(filePath)
					So(statErr, ShouldBeNil)
				})

				err := task.EraseOutput()
				So(err, ShouldBeNil)

				Convey("After eraseOutput file should not exist", func() {
					_, statErr := os.Stat(filePath)
					So(statErr, ShouldNotBeNil)
				})
			})
		})
	})

	Convey("When command which does not exists is executed", func() {
		task, err := executor.Execute("commandThatDoesNotExists")
		So(err, ShouldBeNil)

		defer task.Stop()
		defer task.Clean()
		defer task.EraseOutput()

		Convey("When we wait for the task to terminate", func() {
			task.Wait(0)

			taskState, taskStatus := task.Status()

			Convey("The task should be terminated", func() {
				So(taskState, ShouldEqual, TERMINATED)
			})

			Convey("And the exit status should be 127", func() {
				So(taskStatus.Get().(int), ShouldEqual, 127)
			})

			Convey("And the eraseOutput should clean the stderr file", func() {
				var filePath string
				switch explicitTask := task.(type) {
				case *localTask:
					filePath = explicitTask.stderrFile.Name()
				case *remoteTask:
					filePath = explicitTask.stderrFile.Name()
				default:
					t.Skip("Skipping test: task is neither instance of localTask nor remoteTask")

				}

				task.Clean()
				Convey("Before eraseOutput file should exist", func() {
					_, statErr := os.Stat(filePath)
					So(statErr, ShouldBeNil)
				})

				err := task.EraseOutput()
				So(err, ShouldBeNil)

				Convey("After eraseOutput file should not exist", func() {
					_, statErr := os.Stat(filePath)
					So(statErr, ShouldNotBeNil)
				})
			})
		})
	})

	Convey("When we execute two tasks in the same time", func() {
		task, err := executor.Execute("echo output1")
		task2, err2 := executor.Execute("echo output2")
		So(err, ShouldBeNil)
		So(err2, ShouldBeNil)

		defer task.Stop()
		defer task2.Stop()
		defer task.Clean()
		defer task2.Clean()
		defer task.EraseOutput()
		defer task2.EraseOutput()

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
				reader, readerErr := task.Stdout()
				So(readerErr, ShouldBeNil)
				So(reader, ShouldNotBeNil)

				data, readErr := ioutil.ReadAll(reader)
				So(readErr, ShouldBeNil)
				So(string(data[:]), ShouldStartWith, "output1")

				reader, readerErr = task2.Stdout()
				So(readerErr, ShouldBeNil)
				So(reader, ShouldNotBeNil)

				data, readErr = ioutil.ReadAll(reader)
				So(readErr, ShouldBeNil)
				So(string(data[:]), ShouldStartWith, "output2")

			})

			Convey("Both exit statuses should be 0", func() {
				So(taskStatus1.Get().(int), ShouldEqual, 0)
				So(taskStatus2.Get().(int), ShouldEqual, 0)
			})
		})
	})

	Convey("When command `echo sleep 0` is executed", func() {
		task, err := executor.Execute("echo sleep 0")
		So(err, ShouldBeNil)

		defer task.Stop()
		defer task.Clean()
		defer task.EraseOutput()

		// Wait for the command to execute.
		// TODO(bplotka): Remove the Sleep, since this is prone to errors on different enviroments.
		time.Sleep(100 * time.Millisecond)

		Convey("When we get Status without the Waiting for it", func() {
			taskState, taskStatus := task.Status()

			Convey("And the task should stated that it terminated", func() {
				So(taskState, ShouldEqual, TERMINATED)
			})

			Convey("And the exit status should be 0", func() {
				So(taskStatus.Get().(int), ShouldEqual, 0)
			})
		})
	})
}
