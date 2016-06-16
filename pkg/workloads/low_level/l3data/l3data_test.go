package l3data

import (
	"errors"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestL3Aggressor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	mockedExecutor := new(mocks.Executor)
	mockedTask := new(mocks.TaskHandle)

	Convey("While using l1d aggressor launcher", t, func() {
		const (
			pathToBinary = "test"
			validCommand = "test 86400"
		)

		Convey("Default configuration should be valid", func() {
			config := DefaultL3Config()
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
			duration := time.Duration(-1 * time.Second)
			incorrectConfiguration := Config{
				Path:     pathToBinary,
				Duration: duration,
			}
			launcher := New(mockedExecutor, incorrectConfiguration)

			Convey("Should launcher return error", func() {
				task, err := launcher.Launch()

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "Launcher configuration is invalid.")
				So(task, ShouldBeNil)
			})
		})

	})
}
