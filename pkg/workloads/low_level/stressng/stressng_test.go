package stressng

import (
	"errors"
	"testing"

	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStressng(t *testing.T) {

	mockedExecutor := new(mocks.Executor)
	mockedTask := new(mocks.TaskHandle)

	Convey("While using l1d aggressor launcher", t, func() {
		const (
			validCommand = "stress-ng -foo --bar"
		)

		Convey("Default configuration should be valid", func() {
			launcher := New(
				mockedExecutor,
				"-foo --bar",
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

	})
}
