package l1instruction

import (
	"errors"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestL1dAggressor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	mockedExecutor := new(mocks.Executor)
	mockedTask := new(mocks.TaskHandle)

	Convey("While using l1d aggressor launcher", t, func() {
		const (
			pathToBinary = "test"
			validCommand = "test 2147483648 20"
		)

		Convey("Default configuration should be valid", func() {
			config := DefaultL1iConfig()
			config.Path = pathToBinary
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
		Convey("While using incorrect configuration", func() {
			Convey("With wrong `intensity` value", func() {
				intensity := -1
				iterations := 10
				incorrectConfiguration := Config{
					Path:       pathToBinary,
					Intensity:  intensity,
					Iterations: iterations,
				}
				launcher := New(mockedExecutor, incorrectConfiguration)

				Convey("Should launcher return error", func() {
					task, err := launcher.Launch()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "intensivity value")
					So(task, ShouldBeNil)
				})
			})
			Convey("With wrong `iteration` value", func() {
				intensity := 1
				iterations := -10
				incorrectConfiguration := Config{
					Path:       pathToBinary,
					Intensity:  intensity,
					Iterations: iterations,
				}
				launcher := New(mockedExecutor, incorrectConfiguration)
				Convey("Should launcher return error", func() {
					task, err := launcher.Launch()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "iterations value")
					So(task, ShouldBeNil)
				})
			})
		})
	})
}
