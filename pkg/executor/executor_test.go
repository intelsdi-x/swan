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
		taskHandle, err := executor.Execute("sleep inf")
		So(err, ShouldBeNil)

		defer taskHandle.Stop()
		defer taskHandle.Clean()
		defer taskHandle.EraseOutput()

		Convey("Task should be still running and exitCode should return error", func() {
			taskState := taskHandle.Status()
			So(taskState, ShouldEqual, RUNNING)
			_, err := taskHandle.ExitCode()
			So(err, ShouldNotBeNil)

			stopErr := taskHandle.Stop()
			So(stopErr, ShouldBeNil)
		})

		Convey("When we wait for task termination with the 1ms timeout", func() {
			isTaskTerminated := taskHandle.Wait(1 * time.Microsecond)

			Convey("The timeout appeach and the task should not be terminated", func() {
				So(isTaskTerminated, ShouldBeFalse)
			})

			Convey("Task should be still running and exitCode should return error", func() {
				taskState := taskHandle.Status()
				So(taskState, ShouldEqual, RUNNING)
				_, err := taskHandle.ExitCode()
				So(err, ShouldNotBeNil)

				stopErr := taskHandle.Stop()
				So(stopErr, ShouldBeNil)
			})

			stopErr := taskHandle.Stop()
			So(stopErr, ShouldBeNil)
		})

		Convey("When we stop the task", func() {
			err := taskHandle.Stop()

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("The task should be terminated and the exitCode should "+
				"indicate that task was killed", func() {
				taskState := taskHandle.Status()
				So(taskState, ShouldEqual, TERMINATED)
				exitcode, err := taskHandle.ExitCode()
				So(err, ShouldBeNil)
				// -1 for Local executor.
				// 129 for Remote executor.
				// TODO: Unify exit code constants in next PR.
				So(exitcode, ShouldBeIn, -1, 129)
			})
		})

		Convey("When multiple go routines waits for task termination", func() {
			var wg sync.WaitGroup
			wg.Add(5)
			for i := 0; i < 5; i++ {
				go func() {
					taskHandle.Wait(0)
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

				err := taskHandle.Stop()
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
		taskHandle, err := executor.Execute("echo output")
		So(err, ShouldBeNil)

		defer taskHandle.Stop()
		defer taskHandle.Clean()
		defer taskHandle.EraseOutput()

		Convey("When we wait for the task to terminate", func() {
			taskHandle.Wait(0)

			taskState := taskHandle.Status()

			Convey("The task should be terminated", func() {
				So(taskState, ShouldEqual, TERMINATED)
			})

			Convey("And the exit status should be 0 and output needs to be 'output'", func() {
				exitcode, err := taskHandle.ExitCode()
				So(err, ShouldBeNil)
				So(exitcode, ShouldEqual, 0)

				stdoutFile, stdoutErr := taskHandle.StdoutFile()
				So(stdoutErr, ShouldBeNil)
				So(stdoutFile, ShouldNotBeNil)

				data, readErr := ioutil.ReadAll(stdoutFile)
				So(readErr, ShouldBeNil)
				So(string(data[:]), ShouldStartWith, "output")

			})

			Convey("And the eraseOutput should clean the stdout file", func() {
				stdoutFile, stdoutErr := taskHandle.StdoutFile()
				So(stdoutErr, ShouldBeNil)
				So(stdoutFile, ShouldNotBeNil)

				taskHandle.Clean()
				Convey("Before eraseOutput file should exist", func() {
					_, statErr := os.Stat(stdoutFile.Name())
					So(statErr, ShouldBeNil)
				})

				err := taskHandle.EraseOutput()
				So(err, ShouldBeNil)

				Convey("After eraseOutput file should not exist", func() {
					_, statErr := os.Stat(stdoutFile.Name())
					So(statErr, ShouldNotBeNil)
				})
			})
		})
	})

	Convey("When command which does not exists is executed", func() {
		taskHandle, err := executor.Execute("commandThatDoesNotExists")
		So(err, ShouldBeNil)

		defer taskHandle.Stop()
		defer taskHandle.Clean()
		defer taskHandle.EraseOutput()

		Convey("When we wait for the task to terminate", func() {
			taskHandle.Wait(0)

			taskState := taskHandle.Status()

			Convey("The task should be terminated", func() {
				So(taskState, ShouldEqual, TERMINATED)
			})

			Convey("And the exit status should be 127", func() {

				exitcode, err := taskHandle.ExitCode()

				So(err, ShouldBeNil)
				So(exitcode, ShouldEqual, 127)
			})

			Convey("And the eraseOutput should clean the stderr file", func() {

				taskHandle.Clean()

				stderrFile, err := taskHandle.StderrFile()
				So(err, ShouldBeNil)

				Convey("Before eraseOutput file should exist", func() {
					_, statErr := os.Stat(stderrFile.Name())
					So(statErr, ShouldBeNil)
				})

				err = taskHandle.EraseOutput()
				So(err, ShouldBeNil)

				Convey("After eraseOutput file should not exist", func() {
					_, statErr := os.Stat(stderrFile.Name())
					So(statErr, ShouldNotBeNil)
				})
			})
		})
	})

	Convey("When we execute two tasks in the same time", func() {
		taskHandle, err := executor.Execute("echo output1")
		taskHandle2, err2 := executor.Execute("echo output2")
		So(err, ShouldBeNil)
		So(err2, ShouldBeNil)

		defer taskHandle.Stop()
		defer taskHandle2.Stop()
		defer taskHandle.Clean()
		defer taskHandle2.Clean()
		defer taskHandle.EraseOutput()
		defer taskHandle2.EraseOutput()

		Convey("When we wait for the tasks to terminate", func() {
			taskHandle.Wait(0)
			taskHandle2.Wait(0)

			taskState1 := taskHandle.Status()
			taskState2 := taskHandle2.Status()

			Convey("The tasks should be terminated", func() {
				So(taskState1, ShouldEqual, TERMINATED)
				So(taskState2, ShouldEqual, TERMINATED)
			})

			Convey("The commands stdouts needs to match 'output1' & 'output2'", func() {

				file, err := taskHandle.StdoutFile()
				So(err, ShouldBeNil)
				So(file, ShouldNotBeNil)

				data, readErr := ioutil.ReadAll(file)
				So(readErr, ShouldBeNil)
				So(string(data[:]), ShouldStartWith, "output1")

				file, err = taskHandle2.StdoutFile()
				So(err, ShouldBeNil)
				So(file, ShouldNotBeNil)

				data, readErr = ioutil.ReadAll(file)
				So(readErr, ShouldBeNil)
				So(string(data[:]), ShouldStartWith, "output2")

			})

			Convey("Both exit statuses should be 0", func() {

				exitcode, err := taskHandle.ExitCode()
				So(err, ShouldBeNil)
				So(exitcode, ShouldEqual, 0)

				exitcode, err = taskHandle2.ExitCode()

				So(err, ShouldBeNil)
				So(exitcode, ShouldEqual, 0)
			})
		})
	})

	Convey("When command `echo sleep 0` is executed", func() {
		taskHandle, err := executor.Execute("echo sleep 0")
		So(err, ShouldBeNil)

		defer taskHandle.Stop()
		defer taskHandle.Clean()
		defer taskHandle.EraseOutput()

		// Wait for the command to execute.
		// TODO(bplotka): Remove the Sleep, since this is prone to errors on different enviroments.
		time.Sleep(100 * time.Millisecond)

		Convey("When we get Status without the Waiting for it", func() {

			taskState := taskHandle.Status()

			Convey("And the task should stated that it terminated", func() {
				So(taskState, ShouldEqual, TERMINATED)
			})

			Convey("And the exit status should be 0", func() {

				exitcode, err := taskHandle.ExitCode()

				So(err, ShouldBeNil)
				So(exitcode, ShouldEqual, 0)
			})
		})
	})
}
