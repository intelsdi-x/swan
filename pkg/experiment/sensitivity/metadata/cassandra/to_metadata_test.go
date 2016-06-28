package cassandra

import (
	"testing"
	"time"

	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCassandraModelTransformation(t *testing.T) {
	experimentModel, phasesModel, measurementsModel := prepareCassandraModel()
	mapper := newToMetadata()
	experiment := mapper.transform(experimentModel, phasesModel, measurementsModel)

	Convey("When Cassandra model is transformed into metadata", t, func() {
		Convey("Experiment metadata should be correct", func() {
			So(experiment.ID, ShouldEqual, "experiment")
			So(experiment.LoadDuration, ShouldResemble, time.Second)
			So(experiment.TuningDuration, ShouldResemble, time.Second)
			So(experiment.LCName, ShouldEqual, "launch critical task")
			So(experiment.LGName, ShouldEqual, "load generator")
			So(experiment.RepetitionsNumber, ShouldEqual, 1)
			So(experiment.LoadPointsNumber, ShouldEqual, 2)
			So(experiment.SLO, ShouldEqual, 123)
		})

		Convey("Phases metadata should be correct", func() {
			So(experiment.Phases[0].ID, ShouldEqual, "phase one")
			So(experiment.Phases[0].LCIsolation, ShouldEqual, "Latency critical isolation phase one")
			So(experiment.Phases[0].LCParameters, ShouldEqual, "Latency critical parameters phase one")

			So(experiment.Phases[1].ID, ShouldEqual, "phase two")
			So(experiment.Phases[1].LCIsolation, ShouldEqual, "Latency critical isolation phase two")
			So(experiment.Phases[1].LCParameters, ShouldEqual, "Latency critical parameters phase two")
		})

		Convey("Aggressors metadata should be correct", func() {
			So(experiment.Phases[0].Aggressors, ShouldHaveLength, 2)
			So(experiment.Phases[0].Aggressors[0].Name, ShouldEqual, "first aggressor phase one")
			So(experiment.Phases[0].Aggressors[0].Parameters, ShouldEqual, "first aggressor parameters phase one")
			So(experiment.Phases[0].Aggressors[0].Isolation, ShouldEqual, "first aggressor isolation phase one")
			So(experiment.Phases[0].Aggressors[1].Name, ShouldEqual, "second aggressor phase one")
			So(experiment.Phases[0].Aggressors[1].Parameters, ShouldEqual, "second aggressor parameters phase one")
			So(experiment.Phases[0].Aggressors[1].Isolation, ShouldEqual, "second aggressor isolation phase one")

			So(experiment.Phases[1].Aggressors, ShouldHaveLength, 2)
			So(experiment.Phases[1].Aggressors[0].Name, ShouldEqual, "first aggressor phase two")
			So(experiment.Phases[1].Aggressors[0].Parameters, ShouldEqual, "first aggressor parameters phase two")
			So(experiment.Phases[1].Aggressors[0].Isolation, ShouldEqual, "first aggressor isolation phase two")
			So(experiment.Phases[1].Aggressors[1].Name, ShouldEqual, "second aggressor phase two")
			So(experiment.Phases[1].Aggressors[1].Parameters, ShouldEqual, "second aggressor parameters phase two")
			So(experiment.Phases[1].Aggressors[1].Isolation, ShouldEqual, "second aggressor isolation phase two")
		})

		Convey("Measurements metadata should be correct", func() {
			So(experiment.Phases[0].Measurements, ShouldHaveLength, 2)

			So(experiment.Phases[0].Measurements[0].Load, ShouldEqual, 0.5)
			So(experiment.Phases[0].Measurements[0].LoadPointQPS, ShouldEqual, 303.0)
			So(experiment.Phases[0].Measurements[0].LGParameters, ShouldEqual, "Load generator parameters measurement 1.1")

			So(experiment.Phases[0].Measurements[1].Load, ShouldEqual, 0.7)
			So(experiment.Phases[0].Measurements[1].LoadPointQPS, ShouldEqual, 666.6)
			So(experiment.Phases[0].Measurements[1].LGParameters, ShouldEqual, "Load generator parameters measurement 1.2")

			So(experiment.Phases[1].Measurements, ShouldHaveLength, 1)

			So(experiment.Phases[1].Measurements[0].Load, ShouldEqual, 0.1)
			So(experiment.Phases[1].Measurements[0].LoadPointQPS, ShouldEqual, 0.75)
			So(experiment.Phases[1].Measurements[0].LGParameters, ShouldEqual, "Load generator parameters measurement 2.1")
		})
	})
}

func prepareCassandraModel() (Experiment, []Phase, []Measurement) {
	experiment := Experiment{
		BaseExperiment: metadata.BaseExperiment{
			ID:                "experiment",
			LoadDuration:      time.Second,
			TuningDuration:    time.Second,
			LCName:            "launch critical task",
			LGName:            "load generator",
			RepetitionsNumber: 1,
			LoadPointsNumber:  2,
			SLO:               123,
		},
	}
	phaseOne := Phase{
		BasePhase: metadata.BasePhase{
			ID:           "phase one",
			LCParameters: "Latency critical parameters phase one",
			LCIsolation:  "Latency critical isolation phase one",
		},
		AggressorNames:      []string{"first aggressor phase one", "second aggressor phase one"},
		AggressorParameters: []string{"first aggressor parameters phase one", "second aggressor parameters phase one"},
		AggressorIsolations: []string{"first aggressor isolation phase one", "second aggressor isolation phase one"},
	}
	phaseTwo := Phase{
		BasePhase: metadata.BasePhase{
			ID:           "phase two",
			LCParameters: "Latency critical parameters phase two",
			LCIsolation:  "Latency critical isolation phase two",
		},
		ExperimentID:        "experiment",
		AggressorNames:      []string{"first aggressor phase two", "second aggressor phase two"},
		AggressorParameters: []string{"first aggressor parameters phase two", "second aggressor parameters phase two"},
		AggressorIsolations: []string{"first aggressor isolation phase two", "second aggressor isolation phase two"},
	}
	measurementOneOne := Measurement{
		Measurement: metadata.Measurement{
			Load:         0.5,
			LoadPointQPS: 303.0,
			LGParameters: "Load generator parameters measurement 1.1",
		},
		PhaseID:      "phase one",
		ExperimentID: "experiment",
	}
	measurementOneTwo := Measurement{
		Measurement: metadata.Measurement{
			Load:         0.7,
			LoadPointQPS: 666.6,
			LGParameters: "Load generator parameters measurement 1.2",
		},
		PhaseID:      "phase one",
		ExperimentID: "experiment",
	}
	measurementTwoOne := Measurement{
		Measurement: metadata.Measurement{
			Load:         0.1,
			LoadPointQPS: 0.75,
			LGParameters: "Load generator parameters measurement 2.1",
		},
		PhaseID:      "phase two",
		ExperimentID: "experiment",
	}

	return experiment, []Phase{phaseOne, phaseTwo}, []Measurement{measurementTwoOne, measurementOneTwo, measurementOneOne}
}
