package experiment

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type FakePhase struct {
	ID          string
	Invocations *int
	repetitions int
	FakeResults []float64
}

func (f FakePhase) Name() string {
	return "Test Phase: " + f.ID
}

func (f FakePhase) Repetitions() int {
	return f.repetitions
}

func (f FakePhase) Run() (float64, error) {
	(*f.Invocations)++
	return f.FakeResults[(*f.Invocations)-1], nil
}

func TestExperiment(t *testing.T) {
	Convey("Creating experiment ", t, func() {
		Convey("with allowed variance set to zero should fail", func() {
			conf := ExperimentConfiguration{0, ""}
			_, err := NewExperiment(conf, nil)
			So(err, ShouldNotBeNil)

		})
		Convey("with allowed variance set to negative should fail", func() {
			conf := ExperimentConfiguration{-1, ""}
			_, err := NewExperiment(conf, nil)
			So(err, ShouldNotBeNil)

		})
		Convey("with allowed variance set to positive value and nil phases should fail", func() {
			conf := ExperimentConfiguration{1, ""}
			_, err := NewExperiment(conf, nil)
			So(err, ShouldNotBeNil)

		})
		Convey("with proper configuration and not empty phases should succeed", func() {
			var phases []Phase
			var Invocation int

			conf := ExperimentConfiguration{1, "/tmp"}

			fakePhase := &FakePhase{
				ID:          "Fake phase 01",
				repetitions: 1,
				Invocations: &Invocation,
				FakeResults: []float64{1.0},
			}
			phases = append(phases, fakePhase)

			exp, err := NewExperiment(conf, phases)
			So(exp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			Convey("Experiment Run() with single phase shall succeed", func() {
				err := exp.Run()
				So(err, ShouldBeNil)
				Convey("with repetition set to 1 phase shall be called once", func() {
					So(*fakePhase.Invocations, ShouldEqual, 1)
				})

			})
		})

	})
}
