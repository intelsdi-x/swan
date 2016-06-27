package cassandra

import (
	"fmt"
	"testing"
	"time"

	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCassandraModelTransformation(t *testing.T) {
	Convey("When Cassandra model is transformed I should receive correct metadata model", t, func() {
		experimentModel, phasesModel, measurementsModel := prepareCassandraModel()

		mapper := newCassandraToMetadata()
		experiment := mapper.transform(experimentModel, phasesModel, measurementsModel)

		soExperimentIsCorrect(experiment)
		soPhasesAreCorrect(experiment)
		soMeasurementsAreCorrect(experiment)
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

func soExperimentIsCorrect(experiment metadata.Experiment) {
	So(experiment.ID, ShouldEqual, "experiment")
	So(experiment.LoadDuration, ShouldResemble, time.Second)
	So(experiment.TuningDuration, ShouldResemble, time.Second)
	So(experiment.LcName, ShouldEqual, "launch critical task")
	So(experiment.LgNames, ShouldResemble, []string{"load generator one"})
	So(experiment.RepetitionsNumber, ShouldEqual, 1)
	So(experiment.LoadPointsNumber, ShouldEqual, 2)
	So(experiment.SLO, ShouldEqual, 123)
}

func soPhasesAreCorrect(experiment metadata.Experiment) {
	soPhaseIsCorrect(experiment.Phases[0], "one")
	soPhaseIsCorrect(experiment.Phases[1], "two")
}

func soPhaseIsCorrect(phase metadata.Phase, name string) {
	So(phase.ID, ShouldEqual, fmt.Sprintf("phase %s", name))
	So(phase.LCIsolation, ShouldEqual, fmt.Sprintf("Latency critical isolation phase %s", name))
	So(phase.LCParameters, ShouldEqual, fmt.Sprintf("Latency critical parameters phase %s", name))
	So(phase.AggressorNames, ShouldResemble, []string{fmt.Sprintf("first aggressor phase %s", name), fmt.Sprintf("second aggressor phase %s", name)})
	So(phase.AggressorParameters, ShouldResemble, []string{fmt.Sprintf("first aggressor parameters phase %s", name), fmt.Sprintf("second aggressor parameters phase %s", name)})
	So(phase.AggressorIsolations, ShouldResemble, []string{fmt.Sprintf("first aggressor isolation phase %s", name), fmt.Sprintf("second aggressor isolation phase %s", name)})
}

func soMeasurementsAreCorrect(experiment metadata.Experiment) {
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
}
