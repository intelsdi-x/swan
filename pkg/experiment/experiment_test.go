package experiment

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/experiment/phase/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

func TestExperiment(t *testing.T) {
	Convey("While doing experiment ", t, func() {
		Convey("With proper configuration and empty phases", func() {
			var phases []phase.Phase
			exp, err := NewExperiment("example-experiment1", phases,
				os.TempDir(), Configuration{logrus.ErrorLevel, false, true})

			Convey("Experiment should return with error", func() {
				So(exp, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("With proper configuration and not empty phases", func() {
			var phases []phase.Phase

			mockedPhase := new(mocks.Phase)
			mockedPhase.On("Name").Return("mock-phase01")

			phases = append(phases, mockedPhase)

			exp, err := NewExperiment("example-experiment", phases,
				os.TempDir(), Configuration{logrus.ErrorLevel, false, true})
			So(exp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("While setting ten repetition to phase", func() {
				mockedPhase.On("Run", mock.AnythingOfType("phase.Session")).Return(nil).Times(10)
				mockedPhase.On("Repetitions").Return(int(10))
				mockedPhase.On("Finalize").Return(nil).Once()
				Convey("Experiment should succeed with 10 phase repetitions", func() {
					err := exp.Run()
					So(err, ShouldBeNil)

					mockedPhase.AssertExpectations(t)
				})
			})
		})

		Convey("With invalid configuration stop experiment if error", func() {
			var phases []phase.Phase

			mockedPhase := new(mocks.Phase)
			mockedPhase.On("Name").Return("mock-phase01")

			phases = append(phases, mockedPhase)

			exp, err := NewExperiment("example-experiment", phases,
				os.TempDir(), Configuration{logrus.ErrorLevel, true, true})
			So(exp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("While setting one repetition to phase", func() {
				mockedPhase.On("Run", mock.AnythingOfType("phase.Session")).Return(errors.New("Production task can't be launched")).Once()
				mockedPhase.On("Repetitions").Return(int(10))
				Convey("Experiment should fail with error phase", func() {
					err := exp.Run()
					So(err, ShouldNotBeNil)

					mockedPhase.AssertExpectations(t)
				})
			})
		})

		Convey("With invalid configuration and not empty phases", func() {
			var phases []phase.Phase

			mockedPhase := new(mocks.Phase)
			mockedPhase.On("Name").Return("mock-phase01")

			phases = append(phases, mockedPhase)

			exp, err := NewExperiment("example-experiment", phases,
				os.TempDir(), Configuration{logrus.ErrorLevel, false, true})
			So(exp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("While setting one repetition to phase", func() {
				mockedPhase.On("Run", mock.AnythingOfType("phase.Session")).Return(errors.New("Production task can't be launched"))
				mockedPhase.On("Repetitions").Return(int(10))
				mockedPhase.On("Finalize").Return(nil)
				Convey("Experiment should fail with error phase", func() {
					err := exp.Run()
					So(err, ShouldBeNil)

					mockedPhase.AssertExpectations(t)
				})
			})
		})
	})
}
