// +build integration

package integration

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

const (
	netstatCommand       = "echo stats | nc -w 1 127.0.0.1 11211"
	defaultMemcachedPath = "workloads/data_caching/memcached/memcached-1.4.25/build/memcached"
	swanPkg              = "github.com/intelsdi-x/swan"
)

// TestMemcachedWithExecutor is an integration test with local executor.
// See README for setup items.
func TestMemcachedWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	// Get optional custom Memcached path from MEMCACHED_PATH.
	memcachedPath := os.Getenv("MEMCACHED_BIN")

	if memcachedPath == "" {
		// If custom path does not exists use default path for built memcached.
		memcachedPath = path.Join(os.Getenv("GOPATH"), "src", swanPkg, defaultMemcachedPath)
	}

	Convey("While using Local Shell in Memcached launcher", t, func() {
		l := executor.NewLocal()
		memcachedLauncher := memcached.New(
			l, memcached.DefaultMemcachedConfig(memcachedPath))

		Convey("When memcached is launched", func() {
			// NOTE: It is needed for memcached to have default port available.
			task, err := memcachedLauncher.Launch()
			if task != nil {
				defer task.Stop()
				defer task.Clean()
				defer task.EraseOutput()
			}

			Convey("There should be no error", func() {
				stopErr := task.Stop()

				So(err, ShouldBeNil)
				So(stopErr, ShouldBeNil)
			})

			Convey("Wait 1 second for memcached to init", func() {
				isTerminated := task.Wait(1 * time.Second)

				Convey("Memcached should be still running", func() {
					stopErr := task.Stop()

					// NOTE: Here you will be failing if the memcached
					// can start because it needs to have default port available.
					So(isTerminated, ShouldBeFalse)

					So(stopErr, ShouldBeNil)
				})

				Convey("When we check the memcached endpoint for stats after 1 second", func() {

					netstatTask, netstatErr := l.Execute(netstatCommand)
					if netstatTask != nil {
						defer netstatTask.Stop()
						defer netstatTask.Clean()
						defer netstatTask.EraseOutput()
					}
					Convey("There should be no error", func() {
						task.Stop()
						netstatTask.Stop()

						So(netstatErr, ShouldBeNil)

					})

					Convey("When we wait for netstat ", func() {
						netstatTask.Wait(0)

						Convey("The netstat task should be terminated, the task status should be 0"+
							" and output resultes with a STAT information", func() {
							netstatTaskState := netstatTask.GetStatus()
							So(netstatTaskState, ShouldEqual, executor.TERMINATED)

							exitCode, err := netstatTask.GetExitCode()
							So(err, ShouldBeNil)
							So(exitCode, ShouldEqual, 0)

							stdoutReader, stdoutErr := netstatTask.Stdout()
							So(stdoutErr, ShouldBeNil)
							So(stdoutReader, ShouldNotBeNil)

							data, readErr := ioutil.ReadAll(stdoutReader)
							So(readErr, ShouldBeNil)
							So(string(data[:]), ShouldStartWith, "STAT")
						})
					})
				})

				Convey("When we stop the memcached task", func() {
					err := task.Stop()

					Convey("There should be no error", func() {
						So(err, ShouldBeNil)
					})

					Convey("The task should be terminated and the task status "+
						"should be -1 or 0", func() {
						taskState := task.GetStatus()
						So(taskState, ShouldEqual, executor.TERMINATED)

						exitCode, err := task.GetExitCode()
						So(err, ShouldBeNil)
						// Memcached on CentOS returns 0 (successful code) after SIGTERM.
						So(exitCode, ShouldBeIn, -1, 0)
					})
				})
			})
		})
	})
}
