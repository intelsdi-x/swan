package memcached

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// TestMemcachedWithMockedExecutor runs a Memcached launcher with the mocked executor to simulate
// different cases like proper process execution and error case.
func TestMemcachedWithMockedExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	const (
		expectedCommand = "test -p 11211 -u memcached -t 4 -m 64 -c 1024"
	)

	mockedExecutor := new(mocks.Executor)
	mockedTaskHandle := new(mocks.TaskHandle)

	Convey("While using Memcached launcher", t, func() {
		memcachedLauncher := New(
			mockedExecutor,
			DefaultMemcachedConfig("test"))
		Convey("While simulating proper execution", func() {
			mockedExecutor.On("Execute", expectedCommand).Return(mockedTaskHandle, nil).Once()

			Convey("Build command should create proper command", func() {
				command := memcachedLauncher.buildCommand()
				So(command, ShouldEqual, expectedCommand)

				Convey("Arguments passed to Executor should be a proper command", func() {
					task, err := memcachedLauncher.Launch()
					So(err, ShouldBeNil)
					So(task, ShouldEqual, mockedTaskHandle)

					mockedExecutor.AssertExpectations(t)
				})
			})

		})

		Convey("While simulating error execution", func() {
			mockedExecutor.On("Execute", expectedCommand).Return(nil, errors.New("test")).Once()

			Convey("Build command should create proper command", func() {
				command := memcachedLauncher.buildCommand()
				So(command, ShouldEqual, expectedCommand)

				Convey("Arguments passed to Executor should be a proper command", func() {
					task, err := memcachedLauncher.Launch()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "test")

					So(task, ShouldBeNil)

					mockedExecutor.AssertExpectations(t)
				})
			})

		})

	})
}
