package executor

import (
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParallel(t *testing.T) {
	Convey("When using Parallel to decorate local executor", t, func() {
		parallel := executor.NewLocalIsolated(executor.NewParallel(5))
		Convey("Process should be executed 5 times", func() {
			task, err := parallel.Execute("sleep inf")

			So(err, ShouldBeNil)
			So(task, ShouldNotBeNil)
			// NOTE: We have to wait a bit for parallel to launch commands, though.
			isStopped := task.Wait(1000 * time.Millisecond)
			So(isStopped, ShouldBeFalse)

			defer task.Stop()
			defer task.Clean()
			defer task.EraseOutput()

			cmd := exec.Command("pgrep", "sleep")
			output, err := cmd.CombinedOutput()
			So(err, ShouldBeNil)

			pids := strings.Split(strings.TrimSpace(string(output)), "\n")
			So(len(pids), ShouldBeGreaterThan, 0)
			Convey("When I stop parallel process", func() {
				err = task.Stop()

				So(err, ShouldBeNil)
				Convey("All the child processes should be stopped", func() {
					isStopped := task.Wait(0)
					So(isStopped, ShouldBeTrue)
					cmd = exec.Command("pgrep", "sleep")
					err = cmd.Run()

					So(err, ShouldNotBeNil)
					So(cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus(), ShouldEqual, 1)
				})
			})
		})
	})
}
