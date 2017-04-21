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

	Convey("While using stress-ng aggressor launcher", t, func() {

		Convey("Default configuration should be valid", func() {
			const validCommand = "stress-ng -foo --bar"
			launcher := New(
				mockedExecutor,
				"stress-ng",
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

		Convey("and for specific aggressors we got command as expected", func() {

			Convey("for new stream based aggressor", func() {
				launcher := NewStream(mockedExecutor)
				So(launcher.Name(), ShouldEqual, "stress-ng-stream")
				mockedExecutor.On("Execute", "stress-ng --stream=1").Return(mockedTask, nil).Once()
				_, err := launcher.Launch()
				So(err, ShouldBeNil)
				mockedExecutor.AssertExpectations(t)

			})

			Convey("for new l1 intensive aggressor", func() {
				launcher := NewCacheL1(mockedExecutor)
				So(launcher.Name(), ShouldEqual, "stress-ng-cache-l1")
				mockedExecutor.On("Execute", "stress-ng --cache=1 --cache-level=1").Return(mockedTask, nil).Once()
				_, err := launcher.Launch()
				So(err, ShouldBeNil)
				mockedExecutor.AssertExpectations(t)

			})

			Convey("for new l3 intensive aggressor", func() {
				launcher := NewCacheL3(mockedExecutor)
				So(launcher.Name(), ShouldEqual, "stress-ng-cache-l3")
				mockedExecutor.On("Execute", "stress-ng --cache=1 --cache-level=3").Return(mockedTask, nil).Once()
				_, err := launcher.Launch()
				So(err, ShouldBeNil)
				mockedExecutor.AssertExpectations(t)

			})

			Convey("for new memcpy aggressor", func() {
				launcher := NewMemCpy(mockedExecutor)
				So(launcher.Name(), ShouldEqual, "stress-ng-memcpy")
				mockedExecutor.On("Execute", "stress-ng --memcpy=1").Return(mockedTask, nil).Once()
				_, err := launcher.Launch()
				So(err, ShouldBeNil)
				mockedExecutor.AssertExpectations(t)

			})

			Convey("for new custom aggressor", func() {
				launcher := NewCustom(mockedExecutor)
				So(launcher.Name(), ShouldEqual, "stress-ng-custom ")
				mockedExecutor.On("Execute", "stress-ng ").Return(mockedTask, nil).Once()
				_, err := launcher.Launch()
				So(err, ShouldBeNil)
				mockedExecutor.AssertExpectations(t)

			})
		})

	})
}
