package experiment

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/experiment/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

func TestExperiment(t *testing.T) {
	Convey("While doing experiment ", t, func() {
		Convey("With proper configuration and empty measurements", func() {
			var measurements []Measurement
			exp, err := NewExperiment("example-experiment1", measurements,
				os.TempDir(), logrus.ErrorLevel)

			Convey("Experiment should return with error", func() {
				So(exp, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("With proper configuration and not empty measurements", func() {
			var measurements []Measurement

			mockedMeasurement := new(mocks.Measurement)
			mockedMeasurement.On("Name").Return("mock-measurement01")

			measurements = append(measurements, mockedMeasurement)

			exp, err := NewExperiment("example-experiment", measurements, os.TempDir(), logrus.ErrorLevel)
			So(exp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			Convey("While setting one repetition to measurement", func() {
				mockedMeasurement.On("Run", mock.AnythingOfType("*logrus.Logger")).Return(nil).Times(10)
				mockedMeasurement.On("Repetitions").Return(10)
				mockedMeasurement.On("Finalize").Return(nil).Once()
				Convey("Experiment should succeed with 10 measurement repetitions", func() {
					err := exp.Run()
					So(err, ShouldBeNil)

					mockedMeasurement.AssertExpectations(t)
				})
			})
		})
	})
}
