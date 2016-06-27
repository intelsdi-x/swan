package cassandra

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCassandraModelTransformation(t *testing.T) {
	experimentModel, phasesModel, measurementsModel := prepareCassandraModel()
	mapper := newCassandraToMetadata()
	experiment := mapper.fromCassandraModelToMetadata(experimentModel, phasesModel, measurementsModel)

	Convey("When Cassandra model is transformed into metadata", t, func() {
		Convey("Experiment metadata should be correct", func() {
			So(experiment.ID, ShouldEqual, "experiment")
			So(experiment.LoadDuration, ShouldResemble, time.Second)
			So(experiment.TuningDuration, ShouldResemble, time.Second)
			So(experiment.LcName, ShouldEqual, "launch critical task")
			So(experiment.LgNames, ShouldResemble, []string{"load generator one"})
			So(experiment.RepetitionsNumber, ShouldEqual, 1)
			So(experiment.LoadPointsNumber, ShouldEqual, 2)
			So(experiment.SLO, ShouldEqual, 123)
		})

		Convey("Phases metadata should be correct", func() {
			So(experiment.Phases[0].ID, ShouldEqual, "phase one")
			So(experiment.Phases[0].LCIsolation, ShouldEqual, "Latency critical isolation phase one")
			So(experiment.Phases[0].LCParameters, ShouldEqual, "Latency critical parameters phase one")
			So(experiment.Phases[0].AggressorNames, ShouldResemble, []string{"first aggressor phase one", "second aggressor phase one"})
			So(experiment.Phases[0].AggressorParameters, ShouldResemble, []string{"first aggressor parameters phase one", "second aggressor parameters phase one"})
			So(experiment.Phases[0].AggressorIsolations, ShouldResemble, []string{"first aggressor isolation phase one", "second aggressor isolation phase one"})

			So(experiment.Phases[1].ID, ShouldEqual, "phase two")
			So(experiment.Phases[1].LCIsolation, ShouldEqual, "Latency critical isolation phase two")
			So(experiment.Phases[1].LCParameters, ShouldEqual, "Latency critical parameters phase two")
			So(experiment.Phases[1].AggressorNames, ShouldResemble, []string{"first aggressor phase two", "second aggressor phase two"})
			So(experiment.Phases[1].AggressorParameters, ShouldResemble, []string{"first aggressor parameters phase two", "second aggressor parameters phase two"})
			So(experiment.Phases[1].AggressorIsolations, ShouldResemble, []string{"first aggressor isolation phase two", "second aggressor isolation phase two"})

		})

		Convey("Measurements metadata should be correct", func() {
			So(experiment.Phases[0].Measurements, ShouldHaveLength, 2)

			So(experiment.Phases[0].Measurements[0].Load, ShouldEqual, 0.5)
			So(experiment.Phases[0].Measurements[0].LoadPointQPS, ShouldEqual, 303.0)
			So(experiment.Phases[0].Measurements[0].LGParameters, ShouldResemble, []string{"Load generator parameters measurement 1.1"})

			So(experiment.Phases[0].Measurements[1].Load, ShouldEqual, 0.7)
			So(experiment.Phases[0].Measurements[1].LoadPointQPS, ShouldEqual, 666.6)
			So(experiment.Phases[0].Measurements[1].LGParameters, ShouldResemble, []string{"Load generator parameters measurement 1.2"})

			So(experiment.Phases[1].Measurements, ShouldHaveLength, 1)

			So(experiment.Phases[1].Measurements[0].Load, ShouldEqual, 0.1)
			So(experiment.Phases[1].Measurements[0].LoadPointQPS, ShouldEqual, 0.75)
			So(experiment.Phases[1].Measurements[0].LGParameters, ShouldResemble, []string{"Load generator parameters measurement 2.1"})

		})
	})
}

func prepareCassandraModel() (Experiment, []Phase, []Measurement) {
	experiment := Experiment{
		ID:                "experiment",
		LoadDuration:      time.Second,
		TuningDuration:    time.Second,
		LcName:            "launch critical task",
		LgNames:           []string{"load generator one"},
		RepetitionsNumber: 1,
		LoadPointsNumber:  2,
		SLO:               123,
	}
	phaseOne := Phase{
		ID:                  "phase one",
		LCParameters:        "Latency critical parameters phase one",
		LCIsolation:         "Latency critical isolation phase one",
		AggressorNames:      []string{"first aggressor phase one", "second aggressor phase one"},
		AggressorParameters: []string{"first aggressor parameters phase one", "second aggressor parameters phase one"},
		AggressorIsolations: []string{"first aggressor isolation phase one", "second aggressor isolation phase one"},
	}
	phaseTwo := Phase{
		ExperimentID:        "experiment",
		ID:                  "phase two",
		LCParameters:        "Latency critical parameters phase two",
		LCIsolation:         "Latency critical isolation phase two",
		AggressorNames:      []string{"first aggressor phase two", "second aggressor phase two"},
		AggressorParameters: []string{"first aggressor parameters phase two", "second aggressor parameters phase two"},
		AggressorIsolations: []string{"first aggressor isolation phase two", "second aggressor isolation phase two"},
	}
	measurementOneOne := Measurement{
		PhaseID:      "phase one",
		ExperimentID: "experiment",
		Load:         0.5,
		LoadPointQPS: 303.0,
		LGParameters: []string{"Load generator parameters measurement 1.1"},
	}
	measurementOneTwo := Measurement{
		PhaseID:      "phase one",
		ExperimentID: "experiment",
		Load:         0.7,
		LoadPointQPS: 666.6,
		LGParameters: []string{"Load generator parameters measurement 1.2"},
	}
	measurementTwoOne := Measurement{
		PhaseID:      "phase two",
		ExperimentID: "experiment",
		Load:         0.1,
		LoadPointQPS: 0.75,
		LGParameters: []string{"Load generator parameters measurement 2.1"},
	}

	return experiment, []Phase{phaseOne, phaseTwo}, []Measurement{measurementTwoOne, measurementOneTwo, measurementOneOne}
}
