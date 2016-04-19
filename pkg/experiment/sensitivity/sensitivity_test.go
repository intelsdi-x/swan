package sensitivity

import (
	"errors"
	"github.com/intelsdi-x/swan/pkg/executor"
	executorMocks "github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/intelsdi-x/swan/pkg/workloads"
	workloadMocks "github.com/intelsdi-x/swan/pkg/workloads/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestExperiment(t *testing.T) {
	Convey("While using sensitivity profile experiment", t, func() {
		lcLauncher := new(workloadMocks.Launcher)
		loadGenerator := new(workloadMocks.LoadGenerator)
		configuration := Configuration{
			SLO:             1,
			TuningTimeout:   1,
			LoadDuration:    1,
			LoadPointsCount: 2,
		}
		var aggressors []workloads.Launcher
		aggressor := new(workloadMocks.Launcher)
		aggressors = append(aggressors, aggressor)
		sensitivityExperiment := NewExperiment(configuration, lcLauncher, loadGenerator, aggressors)

		Convey("But production task can't be launched", func() {
			lcLauncher.On("Launch").Return(*new(executor.Task), errors.New("Production task can't be launched"))

			productionTaskLaunchError := sensitivityExperiment.Run()
			So(productionTaskLaunchError.Error(), ShouldEqual, "Production task can't be launched")
			lcLauncher.AssertExpectations(t)
		})

		Convey("And task is launched successfully", func() {
			lcTask := new(executorMocks.Task)
			lcTask.On("Stop").Return(nil)
			lcLauncher.On("Launch").Return(lcTask, nil).Once()

			Convey("But load generator can't be tuned", func() {
				loadGenerator.On("Tune", 1, 1).Return(0, errors.New("Load generator can't be tuned"))

				loadGeneratorTuningError := sensitivityExperiment.Run()
				So(loadGeneratorTuningError.Error(), ShouldEqual, "Load generator can't be tuned")
				loadGenerator.AssertExpectations(t)
				lcLauncher.AssertExpectations(t)
				lcTask.AssertExpectations(t)
			})

			Convey("And load generator is tuned", func() {
				loadGenerator.On("Tune", 1, 1).Return(2, nil)

				Convey("But production task can't be launched during measuring", func() {
					lcLauncher.On("Launch").Return(
						*new(executor.Task),
						errors.New("Production task can't be launched during measuring")).Once()

					loadGeneratorTuningError := sensitivityExperiment.Run()
					So(
						loadGeneratorTuningError.Error(),
						ShouldEqual,
						"Production task can't be launched during measuring")
					loadGenerator.AssertExpectations(t)
					lcLauncher.AssertExpectations(t)
					lcTask.AssertExpectations(t)

				})
				Convey("And production task is launched successfully during measuring", func() {
					lcMeasuringTask := new(executorMocks.Task)
					lcMeasuringTask.On("Stop").Return(nil)
					lcLauncher.On("Launch").Return(lcMeasuringTask, nil)

					Convey("But aggressor can't be launched", func() {
						aggressor.On("Launch").Return(
							*new(executor.Task),
							errors.New("Aggressor task can't be launched"))

						aggressorLaunchError := sensitivityExperiment.Run()
						So(aggressorLaunchError.Error(), ShouldEqual, "Aggressor task can't be launched")
						lcMeasuringTask.AssertExpectations(t)
						lcLauncher.AssertExpectations(t)

					})

					Convey("And aggressor can be launched", func() {
						aggressorTask := new(executorMocks.Task)
						aggressorTask.On("Stop").Return(nil)
						aggressor.On("Launch").Return(aggressorTask, nil)

						Convey("But load testing fails", func() {
							loadGenerator.On("Load", 1, 1).Return(0, errors.New("Load testing fails"))

							aggressorLaunchError := sensitivityExperiment.Run()
							So(aggressorLaunchError.Error(), ShouldEqual, "Load testing fails")
							lcMeasuringTask.AssertExpectations(t)
							lcLauncher.AssertExpectations(t)
							aggressorTask.AssertExpectations(t)
						})

						Convey("And load testing is successful", func() {
							loadGenerator.On("Load", 1, 1).Return(666, nil)

							thereIsNoError := sensitivityExperiment.Run()
							So(thereIsNoError, ShouldBeNil)
							lcMeasuringTask.AssertExpectations(t)
							lcLauncher.AssertExpectations(t)
							aggressorTask.AssertExpectations(t)
						})

					})
				})
			})
		})
	})
}
