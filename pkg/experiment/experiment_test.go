package experiment

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type FakePhase struct {
	ID          string
	Invocations *int
	repetitions int
}

func (f FakePhase) Name() string {
	return f.ID
}

func (f FakePhase) Repetitions() int {
	return f.repetitions
}

func (f FakePhase) Run() error {
	(*f.Invocations)++
	return nil
}

func TestExperiment(t *testing.T) {
	Convey("Creating experiment ", t, func() {
		Convey("With proper configuration and not empty phases should succeed", func() {
			var phases []Phase
			var Invocation int

			fakePhase := &FakePhase{
				ID:          "Fake phase 01",
				repetitions: 1,
				Invocations: &Invocation,
			}
			phases = append(phases, fakePhase)

			exp, err := NewExperiment("example-experiment", phases, "/tmp")
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
