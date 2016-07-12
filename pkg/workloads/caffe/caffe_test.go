package caffe

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/intelsdi-x/testify/mock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vektra/errors"
)

func TestCaffeWithMockedExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("When I create Caffe with mocked executor and default configuration", t, func() {
		mExecutor := new(mocks.Executor)
		mHandle := new(mocks.TaskHandle)

		c := New(mExecutor, DefaultConfig())
		Convey("When I launch the workload with success", func() {
			mExecutor.On("Execute", mock.AnythingOfType("string")).Return(mHandle, nil).Once()
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
			mExecutor.On("Execute", mock.AnythingOfType("string")).Return(nil, expectedErr).Once()
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
			So(config.BinaryPath, ShouldNotBeBlank)
		})
		Convey("Solver field is not blank", func() {
			So(config.SolverPath, ShouldNotBeBlank)
		})
		Convey("Workdir field is not blank", func() {
			So(config.WorkdirPath, ShouldNotBeBlank)
		})
	})
}
