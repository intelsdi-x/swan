package cassandra

import (
	"sort"

	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"
)

// NewCassandraToMetadata creates new instance of struct that transforms Cassandra models to metadata.Experiment.
func newCassandraToMetadata() *cassandraToMetadata {
	return &cassandraToMetadata{phaseNameToIndex: make(map[string]int)}
}

type cassandraToMetadata struct {
	phaseNameToIndex map[string]int
	experiment       metadata.Experiment
}

func (c *cassandraToMetadata) fromCassandraModelToMetadata(experiment Experiment, phases []Phase, measurements []Measurement) metadata.Experiment {
	experimentMetadata := c.buildExperimentMetadataFromModel(experiment)
	experimentMetadata = c.addPhasesToExperiment(experimentMetadata, phases)
	experimentMetadata = c.addMeasurementsToPhases(experimentMetadata, measurements)

	return experimentMetadata
}

func (c *cassandraToMetadata) buildExperimentMetadataFromModel(experimentModel Experiment) metadata.Experiment {
	experimentMetadata := metadata.Experiment{
		BaseExperiment: experimentModel.BaseExperiment,
	}

	return experimentMetadata
}

func (c *cassandraToMetadata) addPhasesToExperiment(experimentMetadata metadata.Experiment, phases []Phase) metadata.Experiment {
	for key, phase := range phases {
		phaseMetadata := metadata.Phase{
			BasePhase: phase.BasePhase,
		}
		phaseMetadata = c.addAggressorsToPhase(phaseMetadata, phase)
		c.phaseNameToIndex[phase.ID+phase.ExperimentID] = key
		experimentMetadata.AddPhase(phaseMetadata)
	}

	return experimentMetadata
}

func (c *cassandraToMetadata) addAggressorsToPhase(phaseMetadata metadata.Phase, phaseModel Phase) metadata.Phase {
	for key, name := range phaseModel.AggressorNames {
		aggressor := metadata.Aggressor{
			Name:       name,
			Parameters: phaseModel.AggressorParameters[key],
			Isolation:  phaseModel.AggressorIsolations[key],
		}
		phaseMetadata.AddAggressor(aggressor)
	}

	return phaseMetadata

}

func (c *cassandraToMetadata) addMeasurementsToPhases(experimentMetadata metadata.Experiment, measurements []Measurement) metadata.Experiment {
	for _, measurement := range measurements {
		phase := &experimentMetadata.Phases[c.phaseNameToIndex[measurement.PhaseID+experimentMetadata.ID]]
		measurementMetadata := metadata.Measurement{
			Load:         measurement.Load,
			LoadPointQPS: measurement.LoadPointQPS,
			LGParameters: measurement.LGParameters,
		}
		phase.Measurements = append(phase.Measurements, measurementMetadata)
	}

	return c.sortMeasurements(experimentMetadata)
}

func (c *cassandraToMetadata) sortMeasurements(experimentMetadata metadata.Experiment) metadata.Experiment {
	for _, phase := range experimentMetadata.Phases {
		sort.Sort(phase.Measurements)
	}

	return experimentMetadata
}
