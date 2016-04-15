package sensitivity

import (
	"testing"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads"
	. "github.com/smartystreets/goconvey/convey"
)

//Implement fake Task interface
type myTask struct {
	state int
}

func (t myTask) Stop() error {
	return nil
}

func (t myTask) Status() (executor.TaskState, *executor.Status) {
	return executor.TERMINATED, &executor.Status{}
}

func (t myTask) Wait(timeoutMs int) bool {
	return true
}

type myLaucher struct {
	status int
}

func (l myLaucher) Launch() (executor.Task, error) {
	return myTask{}, nil
}

// Implement fake LoadGenerator interface
type myLoadGenerator struct {
	status int
}

func (g myLoadGenerator) Tune(slo int, timeoutMs int) (targetQPS int, err error) {
	return 0, nil
}

func (g myLoadGenerator) Load(qps int, durationMs int) (sli int, err error) {
	return 0, nil
}

func TestExperiment(t *testing.T) {
	Convey("Run new Experiment with faked launchers", t, func() {
		var (
			pr      workloads.Launcher
			lgForPr workloads.LoadGenerator
			aggrs   []workloads.Launcher
		)
		conf := Configuration{
			SLO:             100,
			TuningTimeout:   200,
			LoadDuration:    10,
			LoadPointsCount: 19,
		}

		pr = myLaucher{}
		lgForPr = myLoadGenerator{}
		aggrs = append(aggrs, myLaucher{})

		exp := NewExperiment(conf, pr, lgForPr, aggrs)
		Convey("should be successful", func() {
			So(exp.Run(), ShouldBeNil)
		})
	})
}
