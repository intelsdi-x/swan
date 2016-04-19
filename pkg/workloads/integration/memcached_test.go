// +build integration

package integration

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// Before test, make sure you have built memcached in /usr/bin e.g:
// - ln -s SWAN_REPO/workloads/data_caching/memcached/memcached-1.4.25/memcached /usr/bin/memcached
// - user memcached is present.
// - apt-get netstat or yum ...

const (
	memcachedPath = "/usr/bin/memcached"
)

// TestMemcachedWithExecutor is an integration test with local executor.
func TestMemcachedWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell in Memcached launcher", t, func() {
		memcachedLauncher := workloads.NewMemcached(
			executor.NewLocal(),
			workloads.DefaultMemcachedConfig(memcachedPath))

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
