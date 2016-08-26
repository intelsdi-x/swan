package integration

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	. "github.com/smartystreets/goconvey/convey"
)

// TestL1InstructionWithExecutor is an integration test with local executor
// You should build low-level binaries from `github.com/intelsdi-x/swan/workloads/low-level-aggressors/` first
func TestL1InstructionWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell in l1instruction launcher", t, func() {
		l := executor.NewLocal()
		l1InstructionLauncher := l1instruction.New(
			l, l1instruction.DefaultL1iConfig())

		Convey("When l1i binary is launched", func() {
			taskHandle, err := l1InstructionLauncher.Launch()
			if taskHandle != nil {
				defer taskHandle.Stop()
				defer taskHandle.Clean()
				defer taskHandle.EraseOutput()
			}

			Convey("There should be no error", func() {
				stopErr := taskHandle.Stop()

				So(err, ShouldBeNil)
				So(stopErr, ShouldBeNil)
			})

			Convey("L1Instruction should be running", func() {
				So(taskHandle.Status(), ShouldEqual, executor.RUNNING)
			})

			Convey("When we stop the l1i task", func() {
				err := taskHandle.Stop()
				Convey("There should be no error", func() {
					So(err, ShouldBeNil)
				})
				Convey("The task should be terminated and the task status should be -1", func() {
					taskState := taskHandle.Status()
					So(taskState, ShouldEqual, executor.TERMINATED)

					exitCode, err := taskHandle.ExitCode()

					So(err, ShouldBeNil)
					So(exitCode, ShouldEqual, -1)
				})
			})
		})

	})

}
