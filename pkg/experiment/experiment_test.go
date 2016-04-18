package experiment

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExperiment(t *testing.T) {
	Convey("Run new experiment ", t, func() {
		Convey("When result are equal shall pass", func() {
			conf := ExperimentConfiguration{1.5, 5}
			phases := []Phase{
				{"Test01", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test02", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test03", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test04", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test05", func() (float64, error) {
					return 1.0, nil
				}},
			}
			exp, _ := NewExperiment(conf, phases)

			So(exp.Run(), ShouldBeNil)
		})
		Convey("When result exceeds variance test should fail", func() {
			conf := ExperimentConfiguration{1, 5}
			phases := []Phase{
				{"Test01", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test02", func() (float64, error) {
					return 5.0, nil
				}},
				{"Test03", func() (float64, error) {
					return 10.0, nil
				}},
				{"Test04", func() (float64, error) {
					return 15.0, nil
				}},
				{"Test05", func() (float64, error) {
					return 20.0, nil
				}},
			}
			exp, _ := NewExperiment(conf, phases)

			So(exp.Run(), ShouldNotBeNil)
		})
		Convey("When phase func return error test should fail", func() {
			conf := ExperimentConfiguration{1, 5}
			phases := []Phase{
				{"Test01", func() (float64, error) {
					return 1.0, errors.New("Sample error")
				}},
				{"Test02", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test03", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test04", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test05", func() (float64, error) {
					return 1.0, nil
				}},
			}
			exp, _ := NewExperiment(conf, phases)

			So(exp.Run(), ShouldNotBeNil)
		})
		Convey("When variance is negative test should fail", func() {
			conf := ExperimentConfiguration{-1, 5}
			phases := []Phase{
				{"Test01", func() (float64, error) {
					return 1.0, errors.New("Sample error")
				}},
				{"Test02", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test03", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test04", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test05", func() (float64, error) {
					return 1.0, nil
				}},
			}
			Convey("Error should not be nil", func() {
				_, err := NewExperiment(conf, phases)
				So(err, ShouldNotBeNil)
			})
			Convey("Exp should be nil", func() {
				exp, _ := NewExperiment(conf, phases)
				So(exp, ShouldBeNil)
			})
		})
		Convey("When PhaseRepCount is negative test should fail", func() {
			conf := ExperimentConfiguration{1, -1}
			phases := []Phase{
				{"Test01", func() (float64, error) {
					return 1.0, errors.New("Sample error")
				}},
				{"Test02", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test03", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test04", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test05", func() (float64, error) {
					return 1.0, nil
				}},
			}
			Convey("Error should not be nil", func() {
				_, err := NewExperiment(conf, phases)
				So(err, ShouldNotBeNil)
			})
			Convey("Exp should be nil", func() {
				exp, _ := NewExperiment(conf, phases)
				So(exp, ShouldBeNil)
			})
		})
		Convey("When Phases slice is nil test should fail", func() {
			conf := ExperimentConfiguration{1, 5}
			phases := []Phase{}

			Convey("Error should not be nil", func() {
				_, err := NewExperiment(conf, phases)
				So(err, ShouldNotBeNil)
			})
			Convey("Exp should be nil", func() {
				exp, _ := NewExperiment(conf, phases)
				So(exp, ShouldBeNil)
			})
		})
		Convey("When one of phases function is nil test should fail", func() {
			conf := ExperimentConfiguration{1, -1}
			phases := []Phase{
				{"Test01", func() (float64, error) {
					return 1.0, errors.New("Sample error")
				}},
				{"Test02", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test03", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test04", func() (float64, error) {
					return 1.0, nil
				}},
				{"Test05", nil},
			}
			Convey("Error should not be nil", func() {
				_, err := NewExperiment(conf, phases)
				So(err, ShouldNotBeNil)
			})
			Convey("Exp should be nil", func() {
				exp, _ := NewExperiment(conf, phases)
				So(exp, ShouldBeNil)
			})
		})
	})
}
