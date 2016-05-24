package workloads

import (
	"io/ioutil"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	netstatCommand = "echo stats | nc -w 1 127.0.0.1 11211"
)

// TestMemcachedWithExecutor is an integration test with local executor.
// See README for setup items.
func TestMemcachedWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell in Memcached launcher", t, func() {
		l := executor.NewLocal()
		memcachedLauncher := memcached.New(
			l, memcached.DefaultMemcachedConfig())

		Convey("When memcached is launched", func() {
			// NOTE: It is needed for memcached to have default port available.
			taskHandle, err := memcachedLauncher.Launch()
			So(err, ShouldBeNil)
			So(taskHandle, ShouldNotBeNil)
			defer taskHandle.Stop()
			defer taskHandle.Clean()
			defer taskHandle.EraseOutput()

			Convey("There should be no error", func() {
				stopErr := taskHandle.Stop()

				So(err, ShouldBeNil)
				So(stopErr, ShouldBeNil)
			})

			Convey("When we check the memcached endpoint for stats after 1 second", func() {
				netstatTaskHandle, netstatErr := l.Execute(netstatCommand)
				if netstatTaskHandle != nil {
					defer netstatTaskHandle.Stop()
					defer netstatTaskHandle.Clean()
					defer netstatTaskHandle.EraseOutput()
				}
				Convey("There should be no error", func() {
					taskHandle.Stop()
					netstatTaskHandle.Stop()

					So(netstatErr, ShouldBeNil)

				})

				Convey("When we wait for netstat ", func() {
					netstatTaskHandle.Wait(0)

					Convey("The netstat task should be terminated, the task status should be 0"+
						" and output resultes with a STAT information", func() {

						netstatTaskState := netstatTaskHandle.Status()
						So(netstatTaskState, ShouldEqual, executor.TERMINATED)

						exitCode, err := netstatTaskHandle.ExitCode()
						So(err, ShouldBeNil)
						So(exitCode, ShouldEqual, 0)

						stdoutFile, stdoutErr := netstatTaskHandle.StdoutFile()

						So(stdoutErr, ShouldBeNil)
						So(stdoutFile, ShouldNotBeNil)

						data, readErr := ioutil.ReadAll(stdoutFile)
						So(readErr, ShouldBeNil)
						So(string(data[:]), ShouldStartWith, "STAT")
					})
				})
			})

			Convey("When we stop the memcached task", func() {
				err := taskHandle.Stop()

				Convey("There should be no error", func() {
					So(err, ShouldBeNil)
				})

				Convey("The task should be terminated and the task status "+
					"should be -1 or 0", func() {

					taskState := taskHandle.Status()
					So(taskState, ShouldEqual, executor.TERMINATED)

					exitCode, err := taskHandle.ExitCode()

					So(err, ShouldBeNil)
					// Memcached on CentOS returns 0 (successful code) after SIGTERM.
					So(exitCode, ShouldBeIn, -1, 0)
				})
			})
		})
	})
}
