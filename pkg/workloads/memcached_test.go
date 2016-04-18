package workloads

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/mocks"
	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// TestMemcachedBuildCommand
func TestMemcachedBuildCommand(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Memcached launcher", t, func() {
		memcachedLauncher := NewMemcached(
			&mocks.Executor{},
			DefaultMemcachedConfig("test"))

		Convey("Build command should create proper command", func() {
			command := memcachedLauncher.buildCommand()

			So(command, ShouldEqual,
				"test -p 11211 -u memcached -t 4 -m 64 -c 1024")
		})
	})
}

// Before test, make sure you have built memcached in /usr/bin e.g:
// ln -s SWAN_REPO/workloads/data_caching/memcached/memcached-1.4.25/memcached /usr/bin/memcached
const (
	memcachedPath = "/usr/bin/memcached"
)

// TestMemcachedWithExecutor is an integration test with local executor.
func TestMemcachedWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	SkipConvey("While using Local Shell in Memcached launcher", t, func() {
		memcachedLauncher := NewMemcached(
			executor.NewLocal(),
			DefaultMemcachedConfig(memcachedPath))

		Convey("When memcached is launched", func() {
			task, err := memcachedLauncher.Launch()

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)

				task.Stop()
			})

			Convey("When we stop the task", func() {
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
}
