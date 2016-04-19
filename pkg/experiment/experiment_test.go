package experiment

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type DemoPhase struct {
	ID          int
	Invocations int
	repetitions int
}

func (phase *DemoPhase) Name() string     { return "demo phase" + string(phase.ID) }
func (phase *DemoPhase) Repetitions() int { return phase.repetitions }
func (phase *DemoPhase) Run() (*executor.Task, error) {
	phase.Invocations++
	return nil, nil
}

func TestExperiment(t *testing.T) {
	Convey("Creating a new experiment", t, func() {
		experiment := NewExperiment("example experiment 1", Phases{})

		Convey("Should have zero phases", func() {
			So(len(experiment.Phases), ShouldEqual, 0)
		})

		Convey("Adding one phase", func() {
			phase1 := &DemoPhase{
				ID:          1,
				Invocations: 0,
				repetitions: 3,
			}

			experiment.AddPhase(phase1)

			Convey("Total phases should be one", func() {
				So(len(experiment.Phases), ShouldEqual, 1)
			})

			Convey("Running experiment", func() {
				experiment.Run()

				So(phase1.Invocations, ShouldEqual, 3)
			})
		})
	})
}
