package workloads

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// TestMemcachedWithMockedExecutor
func TestMemcachedWithMockedExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	const (
		expectedCommand = "test -p 11211 -u memcached -t 4 -m 64 -c 1024"
	)

	mockedExecutor := new(mocks.Executor)
	mockedExecutor.On("Execute", expectedCommand).Return(nil, nil).Once()

	Convey("While using Memcached launcher", t, func() {
		memcachedLauncher := NewMemcached(
			mockedExecutor,
			DefaultMemcachedConfig("test"))

		Convey("Build command should create proper command", func() {
			command := memcachedLauncher.buildCommand()

			So(command, ShouldEqual, expectedCommand)
		})

		Convey("Arguments passed to Executor should be a proper command", func() {
			memcachedLauncher.Launch()
			mockedExecutor.AssertExpectations(t)
		})
	})
}
