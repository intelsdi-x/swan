package executor

import (
	"bytes"
	"os/exec"
	"syscall"
	"testing"
	"time"

	production "github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParallel(t *testing.T) {
	Convey("When I use Parallel to decorate local executor", t, func() {
		unshare, err := isolation.NewNamespace(syscall.CLONE_NEWPID)
		So(err, ShouldBeNil)

		parallel := production.NewLocalIsolated(isolation.Decorators{production.NewParallel(5), unshare})
		Convey("Process should be executed 5 times", func() {
			task, err := parallel.Execute("sleep inf")

			So(err, ShouldBeNil)
			So(task, ShouldNotBeNil)
			// Yes, this is evil. We have to wait a bit for parallel to launch commands, though.
			isStopped := task.Wait(1000 * time.Millisecond)
			So(isStopped, ShouldBeFalse)

			defer task.Stop()
			defer task.Clean()
			defer task.EraseOutput()

			cmd := exec.Command("pgrep", "sleep")
			output, err := cmd.CombinedOutput()
			So(err, ShouldBeNil)

			// Remove trailing new line
			output = bytes.TrimRight(output, "\n")
			// Split by new line
			pids := bytes.Split(output, []byte{10})
			// In some cases pgrep might not display the most recently created process
			So(len(pids), ShouldBeGreaterThan, 0)
			Convey("When I stop parallel process", func() {
				err = task.Stop()

				So(err, ShouldBeNil)
				Convey("All the child processes should be stopped", func() {
					isStopped := task.Wait(0 * time.Nanosecond)
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
