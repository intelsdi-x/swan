// +build integration

package integration

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
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
		memcachedPath = os.Getenv("GOPATH") + "/src/" + swanPkg + "/" + defaultMemcachedPath
	}

	Convey("While using Local Shell in Memcached launcher", t, func() {
		l := executor.NewLocal()
		memcachedLauncher := memcached.New(
			l, memcached.DefaultMemcachedConfig(memcachedPath))

		Convey("When memcached is launched", func() {
			task, err := memcachedLauncher.Launch()

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				task.Stop()
			})

			Convey("Wait 1 second for memcached to init", func() {
				isTerminated := task.Wait(1)

				Convey("Memcached should be still running", func() {
					So(isTerminated, ShouldBeFalse)

					task.Stop()
				})

				Convey("When we check the memcached endpoint for stats after 1 second", func() {

					netstatTask, netstatErr := l.Execute(netstatCommand)

					Convey("There should be no error", func() {
						So(netstatErr, ShouldBeNil)

						task.Stop()
						netstatTask.Stop()
					})

					Convey("When we wait for netstat ", func() {
						netstatTask.Wait(0)

						Convey("The netstat task should be terminated, the task status should be 0"+
							" and output resultes with a STAT information", func() {
							netstatTaskState, netstatTaskStatus := netstatTask.Status()

							So(netstatTaskState, ShouldEqual, executor.TERMINATED)
							So(netstatTaskStatus.ExitCode, ShouldEqual, 0)
							So(netstatTaskStatus.Stdout, ShouldStartWith, "STAT")

							task.Stop()
							netstatTask.Stop()
						})
					})
				})

				Convey("When we stop the memcached task", func() {
					err := task.Stop()

					Convey("There should be no error", func() {
						So(err, ShouldBeNil)
					})

					Convey("The task should be terminated and the task status should be -1", func() {
						taskState, taskStatus := task.Status()
						So(taskState, ShouldEqual, executor.TERMINATED)
						So(taskStatus.ExitCode, ShouldEqual, -1)
					})
				})
			})
		})
	})
}
