package sensitivity

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	executorMocks "github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/intelsdi-x/swan/pkg/workloads"
	workloadMocks "github.com/intelsdi-x/swan/pkg/workloads/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestSensitivityExperiment(t *testing.T) {
	Convey("While using sensitivity profile experiment", t, func() {
		mockedLcLauncher := new(workloadMocks.Launcher)
		mockedLoadGenerator := new(workloadMocks.LoadGenerator)
		configuration := Configuration{
			SLO:             1,
			LoadDuration:    1 * time.Second,
			LoadPointsCount: 2,
			Repetitions:     1,
		}
		var aggressors []workloads.Launcher
		mockedAggressor := new(workloadMocks.Launcher)
		aggressors = append(aggressors, mockedAggressor)

		sensitivityExperiment, err := InitExperiment("test", logrus.ErrorLevel, configuration,
			mockedLcLauncher, mockedLoadGenerator, aggressors)
		So(err, ShouldBeNil)

		Convey("But production task can't be launched", func() {
			mockedLcLauncher.On("Launch").Return(nil,
				errors.New("Production task can't be launched"))

			productionTaskLaunchError := sensitivityExperiment.Run()
			So(productionTaskLaunchError.Error(), ShouldEqual, "Production task can't be launched")
			mockedLcLauncher.AssertExpectations(t)
		})

		Convey("And task is launched successfully", func() {
			mockedLcTaskHandle := new(executorMocks.TaskHandle)
			mockedLcTaskHandle.On("Stop").Return(nil)
			mockedLcTaskHandle.On("Clean").Return(nil)
			mockedLcLauncher.On("Launch").Return(mockedLcTaskHandle, nil).Once()

			Convey("But load generator can't be tuned", func() {
				mockedLoadGenerator.On("Tune", 1).Return(0, 0, errors.New("Load generator can't be tuned"))

				loadGeneratorTuningError := sensitivityExperiment.Run()
				So(loadGeneratorTuningError.Error(), ShouldEqual, "Load generator can't be tuned")
				mockedLoadGenerator.AssertExpectations(t)
				mockedLcLauncher.AssertExpectations(t)
				mockedLcTaskHandle.AssertExpectations(t)
			})

			Convey("And load generator can be tuned", func() {
				mockedLoadGenerator.On("Tune", 1).Return(2, 2, nil)

				Convey("But production task can't be launched during measuring", func() {
					mockedLcLauncher.On("Launch").Return(nil,
						errors.New("Production task can't be launched during measuring")).Once()

					loadGeneratorTuningError := sensitivityExperiment.Run()
					So(loadGeneratorTuningError.Error(), ShouldEqual,
						"Production task can't be launched during measuring")
					mockedLoadGenerator.AssertExpectations(t)
					mockedLcLauncher.AssertExpectations(t)
					mockedLcTaskHandle.AssertExpectations(t)

				})
				Convey("And production task is launched successfully during measuring", func() {
					mockedLcMeasuringTaskHandle := new(executorMocks.TaskHandle)
					mockedLcMeasuringTaskHandle.On("Stop").Return(nil)
					mockedLcMeasuringTaskHandle.On("Clean").Return(nil)
					mockedLcLauncher.On("Launch").Return(mockedLcMeasuringTaskHandle, nil)
					Convey("But aggressor can't be launched", func() {
						mockedAggressor.On("Launch").Return(
							*new(executor.TaskHandle),
							errors.New("Aggressor task can't be launched"))
						mockedLoadGenerator.On("Load", 1, 1*time.Second).Return(666, 222, nil)
						mockedLoadGenerator.On("Load", 2, 1*time.Second).Return(666, 222, nil)

						aggressorLaunchError := sensitivityExperiment.Run()
						So(aggressorLaunchError.Error(), ShouldEqual, "Aggressor task can't be launched")
						mockedLcMeasuringTaskHandle.AssertExpectations(t)
						mockedLcLauncher.AssertExpectations(t)

					})

					Convey("And aggressor can be launched", func() {
						mockedAggressorTaskHandle := new(executorMocks.TaskHandle)
						mockedAggressor.On("Launch").Return(mockedAggressorTaskHandle, nil)

						Convey("But load testing fails", func() {
							mockedLoadGenerator.On("Load", 1, 1*time.Second).
								Return(0, 0, errors.New("Load testing failed"))

							aggressorLaunchError := sensitivityExperiment.Run()
							So(aggressorLaunchError.Error(), ShouldEqual, "Load testing failed")
							mockedLcMeasuringTaskHandle.AssertExpectations(t)
							mockedLcLauncher.AssertExpectations(t)
							mockedAggressorTaskHandle.AssertExpectations(t)
						})

						Convey("And load testing is successful", func() {
							mockedAggressorTaskHandle.On("Stop").Return(nil)
							mockedAggressorTaskHandle.On("Clean").Return(nil)
							mockedLoadGenerator.On("Load", 1, 1*time.Second).Return(666, 222, nil)
							mockedLoadGenerator.On("Load", 2, 1*time.Second).Return(666, 222, nil)

							thereIsNoError := sensitivityExperiment.Run()
							So(thereIsNoError, ShouldBeNil)
							mockedLcMeasuringTaskHandle.AssertExpectations(t)
							mockedLcLauncher.AssertExpectations(t)
							mockedAggressorTaskHandle.AssertExpectations(t)
						})

					})
				})
			})
		})
	})
}
