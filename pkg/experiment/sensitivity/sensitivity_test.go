package sensitivity

import (
	"errors"
	"github.com/intelsdi-x/swan/pkg/executor"
	executorMocks "github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/intelsdi-x/swan/pkg/workloads"
	workloadMocks "github.com/intelsdi-x/swan/pkg/workloads/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestExperiment(t *testing.T) {
	Convey("While using sensitivity profile experiment", t, func() {
		mockedLcLauncher := new(workloadMocks.Launcher)
		mockedLoadGenerator := new(workloadMocks.LoadGenerator)
		configuration := Configuration{
			SLO:             1,
			LoadDuration:    1 * time.Second,
			LoadPointsCount: 2,
		}
		var aggressors []workloads.Launcher
		mockedAggressor := new(workloadMocks.Launcher)
		aggressors = append(aggressors, mockedAggressor)
		sensitivityExperiment := NewExperiment(configuration, mockedLcLauncher, mockedLoadGenerator, aggressors)

		Convey("But production task can't be launched", func() {
			mockedLcLauncher.On("Launch").Return(*new(executor.Task), errors.New("Production task can't be launched"))

			productionTaskLaunchError := sensitivityExperiment.Run()
			So(productionTaskLaunchError.Error(), ShouldEqual, "Production task can't be launched")
			mockedLcLauncher.AssertExpectations(t)
		})

		Convey("And task is launched successfully", func() {
			mockedLcTask := new(executorMocks.Task)
			mockedLcTask.On("Stop").Return(nil)
			mockedLcLauncher.On("Launch").Return(mockedLcTask, nil).Once()

			Convey("But load generator can't be tuned", func() {
				mockedLoadGenerator.On("Tune", 1).Return(0, 0, errors.New("Load generator can't be tuned"))

				loadGeneratorTuningError := sensitivityExperiment.Run()
				So(loadGeneratorTuningError.Error(), ShouldEqual, "Load generator can't be tuned")
				mockedLoadGenerator.AssertExpectations(t)
				mockedLcLauncher.AssertExpectations(t)
				mockedLcTask.AssertExpectations(t)
			})

			Convey("And load generator is tuned", func() {
				mockedLoadGenerator.On("Tune", 1).Return(2, 2, nil)

				Convey("But production task can't be launched during measuring", func() {
					mockedLcLauncher.On("Launch").Return(
						*new(executor.Task),
						errors.New("Production task can't be launched during measuring")).Once()

					loadGeneratorTuningError := sensitivityExperiment.Run()
					So(
						loadGeneratorTuningError.Error(),
						ShouldEqual,
						"Production task can't be launched during measuring")
					mockedLoadGenerator.AssertExpectations(t)
					mockedLcLauncher.AssertExpectations(t)
					mockedLcTask.AssertExpectations(t)

				})
				Convey("And production task is launched successfully during measuring", func() {
					mockedLcMeasuringTask := new(executorMocks.Task)
					mockedLcMeasuringTask.On("Stop").Return(nil)
					mockedLcLauncher.On("Launch").Return(mockedLcMeasuringTask, nil)

					Convey("But aggressor can't be launched", func() {
						mockedAggressor.On("Launch").Return(
							*new(executor.Task),
							errors.New("Aggressor task can't be launched"))

						aggressorLaunchError := sensitivityExperiment.Run()
						So(aggressorLaunchError.Error(), ShouldEqual, "Aggressor task can't be launched")
						mockedLcMeasuringTask.AssertExpectations(t)
						mockedLcLauncher.AssertExpectations(t)

					})

					Convey("And aggressor can be launched", func() {
						mockedAggressorTask := new(executorMocks.Task)
						mockedAggressorTask.On("Stop").Return(nil)
						mockedAggressor.On("Launch").Return(mockedAggressorTask, nil)

						Convey("But load testing fails", func() {
							mockedLoadGenerator.On("Load", 1, 1*time.Second).
								Return(0, 0, errors.New("Load testing fails"))

							aggressorLaunchError := sensitivityExperiment.Run()
							So(aggressorLaunchError.Error(), ShouldEqual, "Load testing fails")
							mockedLcMeasuringTask.AssertExpectations(t)
							mockedLcLauncher.AssertExpectations(t)
							mockedAggressorTask.AssertExpectations(t)
						})

						Convey("And load testing is successful", func() {
							mockedLoadGenerator.On("Load", 1, 1*time.Second).Return(666, 222, nil)

							thereIsNoError := sensitivityExperiment.Run()
							So(thereIsNoError, ShouldBeNil)
							mockedLcMeasuringTask.AssertExpectations(t)
							mockedLcLauncher.AssertExpectations(t)
							mockedAggressorTask.AssertExpectations(t)
						})

					})
				})
			})
		})
	})
}
