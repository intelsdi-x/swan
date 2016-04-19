// +build integration

package integration

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

const (
	netstatCommand        = "echo stats | nc -q 5 127.0.0.1 11211"
	memcachedRelativePath = "workloads/data_caching/memcached/memcached-1.4.25/memcached"
)

// Before test, make sure you have built memcached in /usr/bin e.g:
// - have $SWAN_ROOT environment variable for Swan repo location.
// - user memcached is present.
// - apt-get netstat or yum ...

// TestMemcachedWithExecutor is an integration test with local executor.
func TestMemcachedWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	memcachedPath := os.Getenv("SWAN_ROOT")

	if memcachedPath == "" {
		t.Fatal("There is no $SWAN_ROOT environment variable for Swan repo location.")
	}

	Convey("While using Local Shell in Memcached launcher", t, func() {
		l := executor.NewLocal()
		memcachedLauncher := workloads.NewMemcached(
			l, workloads.DefaultMemcachedConfig(memcachedPath+"/"+memcachedRelativePath))

		Convey("When memcached is launched", func() {
			task, err := memcachedLauncher.Launch()

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				task.Stop()
			})

			Convey("When we check the memcached endpoint for stats", func() {
				netstatTask, netstatErr := l.Execute(netstatCommand)

				Convey("There should be no error", func() {
					So(netstatErr, ShouldBeNil)

					task.Stop()
					netstatTask.Stop()
				})

				Convey("When we wait for netstat ", func() {
					netstatTask.Wait(0)

					Convey("The netstat task should be terminated, the task status should be 0"+
						" and output resultes with a STATA information", func() {
						netstatTaskState, netstatTaskStatus := netstatTask.Status()

						So(netstatTaskState, ShouldEqual, executor.TERMINATED)
						So(netstatTaskStatus.ExitCode, ShouldEqual, 0)
						So(netstatTaskStatus.Stdout, ShouldStartWith, "STAT")
					})

					task.Stop()
					netstatTask.Stop()
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
