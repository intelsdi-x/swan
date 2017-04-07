package stream

import (
	"errors"
	"testing"

	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStreamAggressor(t *testing.T) {
	// log.SetLevel(log.ErrorLevel)

	mockedExecutor := new(mocks.Executor)
	mockedTask := new(mocks.TaskHandle)

	Convey("While using stream aggressor launcher", t, func() {

		Convey("when using default configuration", func() {
			const validCommand = "env OMP_NUM_THREADS=0 test1"
			config := DefaultConfig()
			config.Path = "test1"
			launcher := New(
				mockedExecutor,
				config,
			)

			Convey("When executor is able to run this command then it should return mocked taskHandle "+
				"without error", func() {

				mockedExecutor.On("Execute", validCommand).Return(mockedTask, nil).Once()

				task, err := launcher.Launch()
				So(err, ShouldBeNil)
				So(task, ShouldEqual, mockedTask)

				mockedExecutor.AssertExpectations(t)
			})
			Convey("When executor isn't able to run this command then it should return error without "+
				"mocked taskHandle", func() {

				mockedExecutor.On("Execute", validCommand).Return(nil, errors.New("fail to execute")).Once()

				task, err := launcher.Launch()
				So(task, ShouldBeNil)
				So(err.Error(), ShouldEqual, "fail to execute")

				mockedExecutor.AssertExpectations(t)
			})
		})

		Convey("when with configuration with threads explicitly", func() {
			const validCommand = "env OMP_NUM_THREADS=5 test2"
			config := DefaultConfig()
			config.Path = "test2"
			config.NumThreads = 5
			launcher := New(
				mockedExecutor,
				config,
			)

			Convey("When executor is able to run this command then it should return mocked taskHandle "+
				"without error", func() {

				mockedExecutor.On("Execute", validCommand).Return(mockedTask, nil).Once()

				task, err := launcher.Launch()
				So(err, ShouldBeNil)
				So(task, ShouldEqual, mockedTask)

				mockedExecutor.AssertExpectations(t)
			})
		})

	})
}
