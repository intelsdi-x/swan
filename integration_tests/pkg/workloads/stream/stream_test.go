package integration

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/executor"
	stream "github.com/intelsdi-x/swan/pkg/workloads/low_level/stream"
	. "github.com/smartystreets/goconvey/convey"
)

// TestStreamWithExecutor is an integration test with local executor.
func TestStreamWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell in stream launcher", t, func() {
		l := executor.NewLocal()
		streamLauncher := stream.New(l, stream.DefaultConfig())

		Convey("When stream binary is launched", func() {
			taskHandle, err := streamLauncher.Launch()
			Convey("task should launch successfully", func() {
				So(err, ShouldBeNil)
				Reset(func() {
					taskHandle.Stop()
					taskHandle.Clean()
					taskHandle.EraseOutput()
				})
				Convey("and stream should be running", func() {
					So(taskHandle.Status(), ShouldEqual, executor.RUNNING)
					Convey("When we stop the stream task", func() {
						err := taskHandle.Stop()
						Convey("There should be no error", func() {
							So(err, ShouldBeNil)
							Convey("and task should be terminated and the task exit status should be -1 (killed)", func() {
								taskState := taskHandle.Status()
								So(taskState, ShouldEqual, executor.TERMINATED)

								exitCode, err := taskHandle.ExitCode()

								So(err, ShouldBeNil)
								So(exitCode, ShouldEqual, -1)
							})
						})
					})
				})
			})

		})

	})

}
