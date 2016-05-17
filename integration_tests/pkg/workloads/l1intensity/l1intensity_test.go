package integration

import (
	"os"
	"path"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1intesity"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	swanPkg                = "github.com/intelsdi-x/swan"
	defaultL1IntensityPath = "workloads/low-level-aggressors/l1i"
)

// TestL1IntensityWithExecutor is an integration test with local executor
// You should build low-level binaries from `github.com/intelsdi-x/swan/workloads/low-level-aggressors/` first
func TestL1IntensityWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	l1IntensityPath := path.Join(os.Getenv("GOPATH"), "src", swanPkg, defaultL1IntensityPath)

	Convey("While using Local Shell in L1Intesity launcher", t, func() {
		l := executor.NewLocal()
		l1IntensityLauncher := l1intesity.New(l, l1intesity.DefaultL1iConfig(l1IntensityPath))

		Convey("When l1i binary is launched", func() {
			taskHandle, err := l1IntensityLauncher.Launch()
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

			Convey("L1Intensity should be running", func() {
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
