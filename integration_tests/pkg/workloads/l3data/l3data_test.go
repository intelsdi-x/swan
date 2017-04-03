package integration

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	. "github.com/smartystreets/goconvey/convey"
)

// TestL3dataWithExecutor is an integration test with local executor
// You should build low-level binaries from `github.com/intelsdi-x/swan/workloads/low-level-aggressors/` first
func TestL3dataWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell in L3Data launcher", t, func() {
		l := executor.NewLocal()
		l3dataLauncher := l3data.New(
			l, l3data.DefaultL3Config())

		Convey("When l3d binary is launched", func() {
			taskHandle, err := l3dataLauncher.Launch()
			if taskHandle != nil {
				defer taskHandle.Stop()
				defer taskHandle.EraseOutput()
			}

			Convey("There should be no error", func() {
				stopErr := taskHandle.Stop()

				So(err, ShouldBeNil)
				So(stopErr, ShouldBeNil)
			})

			Convey("L3data should be running", func() {
				So(taskHandle.Status(), ShouldEqual, executor.RUNNING)
			})

			Convey("When we stop the l3d task", func() {
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
