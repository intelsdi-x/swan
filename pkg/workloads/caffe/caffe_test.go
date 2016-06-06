package caffe

import (
	"testing"

	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vektra/errors"
)

// TestMemcachedWithMockedExecutor runs a Memcached launcher with the mocked executor to simulate
// different cases like proper process execution and error case.
func TestCaffeWithMockedExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("When I create Caffe with mocked executor and default configuration", t, func() {
		mExecutor := new(mocks.Executor)
		mHandle := new(mocks.TaskHandle)

		c := New(mExecutor, DefaultConfig())
		expectedCommand := fmt.Sprintf("%s train --solver=%s",
			DefaultConfig().PathToBinary,
			DefaultConfig().PathToSolver)

		Convey("When I launch the workload with success", func() {
			mExecutor.On("Execute", expectedCommand).Return(mHandle, nil).Once()
			handle, err := c.Launch()
			Convey("Proper handle is returned", func() {
				So(handle, ShouldEqual, mHandle)
			})
			Convey("Error is nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I launch the workload with failure", func() {
			expectedErr := errors.New("example error")
			mExecutor.On("Execute", expectedCommand).Return(nil, expectedErr).Once()
			handle, err := c.Launch()
			Convey("Proper handle is returned", func() {
				So(handle, ShouldBeNil)
			})
			Convey("Error is nil", func() {
				So(err, ShouldEqual, expectedErr)
			})
		})
	})
}

func TestCaffeDefaultConfig(t *testing.T) {
	Convey("When I create default config for Caffe", t, func() {
		config := DefaultConfig()
		Convey("Binary field is not blank", func() {
			So(config.PathToBinary, ShouldNotBeBlank)
		})
		Convey("Solver field is not blank", func() {
			So(config.PathToSolver, ShouldNotBeBlank)
		})
	})
}
