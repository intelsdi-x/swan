package sensitivity

import (
	"testing"
	"github.com/intelsdi-x/swan/pkg/executor"
	executorMocks "github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/intelsdi-x/swan/pkg/workloads"
	workloadMocks "github.com/intelsdi-x/swan/pkg/workloads/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"errors"
)

func TestExperiment(t *testing.T) {
	Convey("While using sensitivity profile experiment", t, func() {
		lcLauncher := new(workloadMocks.Launcher)
		loadGenerator := new(workloadMocks.LoadGenerator)
		configuration := Configuration{
			SLO: 1,
			TuningTimeout: 1,
			LoadDuration: 1,
			LoadPointsCount: 1,
		}
		var aggressors []workloads.Launcher
		aggressors = append(aggressors, new(workloadMocks.Launcher))
		sensitivityExperiment := NewExperiment(configuration, lcLauncher, loadGenerator, aggressors)

		Convey("But production task can't be launched", func() {
			lcLauncher.On("Launch").Return(*new(executor.Task), errors.New("Production task can't be launched"))
			loadGenerator.AssertNotCalled(t, "Tune", 1, 1)

			productionTaskLaunchError := sensitivityExperiment.Run()
			So(productionTaskLaunchError.Error(), ShouldEqual, "Production task can't be launched")
		})

		Convey("And task is started successfully", func() {
			lcTask := new(executorMocks.Task)
			lcTask.On("Stop").Return(nil)
			lcLauncher.On("Launch").Return(lcTask, nil)

			Convey("But load generator can't be tuned", func() {
				loadGenerator.On("Tune", 1, 1).Return(0, errors.New("Load generator can't be tuned"))

				loadGeneratorTuningError := sensitivityExperiment.Run()
				So(loadGeneratorTuningError.Error(), ShouldEqual, "Load generator can't be tuned")
			})
		})
	})
}
