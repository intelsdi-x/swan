// +build sequential parallel

package executor

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	. "github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
)

// testExecutor tests the execution of process for given executor.
// This test can be used inside any Executor implementation test.
func testExecutor(t *testing.T, executor Executor) {
	Convey("When blocking infinitively sleep command is executed", func() {
		taskHandle, err := executor.Execute("sleep inf")
		So(err, ShouldBeNil)

		defer StopCleanAndErase(taskHandle)

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
				// 137 for Remote executor (process killed).
				// TODO: Unify exit code constants in next PR.
				So(exitcode, ShouldBeIn, -1, 137)
			})
		})
	})

	Convey("When command `echo output` is executed", func() {
		taskHandle, err := executor.Execute("echo output")
		So(err, ShouldBeNil)

		defer StopCleanAndErase(taskHandle)

		Convey("When we wait for the task to terminate. The exit status should be 0 and output needs to be 'output'", func() {
			So(taskHandle.Wait(0), ShouldBeTrue)
			taskState := taskHandle.Status()
			So(taskState, ShouldEqual, TERMINATED)

			exitcode, err := taskHandle.ExitCode()
			So(err, ShouldBeNil)
			So(exitcode, ShouldEqual, 0)

			stdoutFile, stdoutErr := taskHandle.StdoutFile()
			So(stdoutErr, ShouldBeNil)
			So(stdoutFile, ShouldNotBeNil)

			// NOTE: the fd may point to the end of the file.
			stdoutFile.Seek(0, 0)

			data, readErr := ioutil.ReadAll(stdoutFile)
			So(readErr, ShouldBeNil)
			// ShouldContain is required because kubernetes pod exectuors adds empty line upfront.
			So(string(data[:]), ShouldContainSubstring, "output")
		})

		Convey("And the eraseOutput should clean the stdout file", func() {
			So(taskHandle.Wait(0), ShouldBeTrue)
			taskState := taskHandle.Status()
			So(taskState, ShouldEqual, TERMINATED)

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

	Convey("When command which does not exists is executed", func() {
		taskHandle, err := executor.Execute("/bin/sh -c commandThatDoesNotExists")
		So(err, ShouldBeNil)

		defer StopCleanAndErase(taskHandle)

		Convey("When we wait for the task to terminate and the exit status should be 127", func() {
			So(taskHandle.Wait(0), ShouldBeTrue)

			taskState := taskHandle.Status()

			So(taskState, ShouldEqual, TERMINATED)

			exitcode, err := taskHandle.ExitCode()

			So(err, ShouldBeNil)
			So(exitcode, ShouldEqual, 127)
		})

		Convey("And the eraseOutput should clean the stderr file", func() {
			So(taskHandle.Wait(0), ShouldBeTrue)
			taskState := taskHandle.Status()
			So(taskState, ShouldEqual, TERMINATED)

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

	Convey("When we execute two tasks in the same time", func() {
		taskHandle, err := executor.Execute("echo output1")
		taskHandle2, err2 := executor.Execute("echo output2")
		So(err, ShouldBeNil)
		So(err2, ShouldBeNil)
		defer StopCleanAndErase(taskHandle)
		defer StopCleanAndErase(taskHandle2)

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
				So(string(data[:]), ShouldContainSubstring, "output1")

				file, err = taskHandle2.StdoutFile()
				So(err, ShouldBeNil)
				So(file, ShouldNotBeNil)

				data, readErr = ioutil.ReadAll(file)
				So(readErr, ShouldBeNil)
				So(string(data[:]), ShouldContainSubstring, "output2")

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
		defer StopCleanAndErase(taskHandle)

		// Wait for the command to execute.
		// TODO(bplotka): Remove the Sleep/Wait, since this is prone to errors on different environments.
		taskHandle.Wait(1 * time.Second)

		Convey("When we get Status with Wait for max 1 second", func() {
			taskState := taskHandle.Status()

			Convey("And the task should stated that it terminated", func() {
				So(taskState, ShouldEqual, TERMINATED)
			})

			Convey("And the exit status should be 0", func() {
				exitcode, err := taskHandle.ExitCode()

				So(err, ShouldBeNil)
				So(exitcode, ShouldEqual, 0)
			})

			Convey("And the output files shall remain", func() {
				stdoutFile, err := taskHandle.StdoutFile()
				So(err, ShouldBeNil)
				stderrFile, err := taskHandle.StderrFile()
				So(err, ShouldBeNil)
				stdoutStat, stdoutErr := os.Stat(stdoutFile.Name())
				stderrStat, stderrErr := os.Stat(stderrFile.Name())
				So(stdoutErr, ShouldBeNil)
				So(stderrErr, ShouldBeNil)
				So(stdoutStat.Mode().IsRegular(), ShouldBeTrue)
				So(stderrStat.Mode().IsRegular(), ShouldBeTrue)
			})
		})
	})

	Convey("When command `sleep 0` is executed and EraseOutput is called output files shall be removed", func() {
		taskHandle, err := executor.Execute("sleep 0")
		So(err, ShouldBeNil)

		taskHandle.Wait(1 * time.Second)

		stdoutFile, _ := taskHandle.StdoutFile()
		stderrFile, _ := taskHandle.StderrFile()

		outputDir, _ := path.Split(stdoutFile.Name())

		taskHandle.Stop()
		taskHandle.Clean()
		taskHandle.EraseOutput()

		_, stdoutErr := os.Stat(stdoutFile.Name())
		_, stderrErr := os.Stat(stderrFile.Name())
		_, outputDirErr := os.Stat(outputDir)

		So(stdoutErr, ShouldNotBeNil)
		So(stderrErr, ShouldNotBeNil)
		So(outputDirErr, ShouldNotBeNil)
	})
}
