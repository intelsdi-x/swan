package experiment

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/experiment/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestExperiment(t *testing.T) {
	Convey("While doing experiment ", t, func() {
		Convey("With proper configuration and empty phases", func() {
			var phases []Phase
			exp, err := NewExperiment("example-experiment1", phases,
				os.TempDir(), logrus.ErrorLevel)

			Convey("Experiment should return with error", func() {
				So(exp, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("With proper configuration and not empty phases", func() {
			var phases []Phase

			mockedPhase := new(mocks.Phase)
			mockedPhase.On("Name").Return("mock-phase01")

			phases = append(phases, mockedPhase)

			exp, err := NewExperiment("example-experiment", phases, os.TempDir(), logrus.ErrorLevel)
			So(exp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("While setting one repetition to phase", func() {
				mockedPhase.On("Run").Return(nil).Times(10)
				mockedPhase.On("Repetitions").Return(10)
				mockedPhase.On("Finalize").Return(nil).Once()
				Convey("Experiment should succeed with 10 phase repetitions", func() {
					err := exp.Run()
					So(err, ShouldBeNil)

					mockedPhase.AssertExpectations(t)
				})
			})
		})
	})
}
