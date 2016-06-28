package cassandra

import (
	"sort"

	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"
)

// NewToMetadata creates new instance of struct that transforms Cassandra models to metadata.Experiment.
func NewToMetadata() *toMetadata {
	return &toMetadata{phaseNameToIndex: make(map[string]int)}
}

type toMetadata struct {
	phaseNameToIndex map[string]int
	experiment       metadata.Experiment
}

func (c *toMetadata) transform(experiment Experiment, phases []Phase, measurements []Measurement) metadata.Experiment {
	experimentMetadata := c.buildExperimentMetadataFromModel(experiment)
	experimentMetadata = c.addPhasesToExperiment(experimentMetadata, phases)
	experimentMetadata = c.addMeasurementsToPhases(experimentMetadata, measurements)

	return experimentMetadata
}

func (c *toMetadata) buildExperimentMetadataFromModel(experimentModel Experiment) metadata.Experiment {
	return metadata.Experiment{
		BaseExperiment: experimentModel.BaseExperiment,
	}
}

func (c *toMetadata) addPhasesToExperiment(experimentMetadata metadata.Experiment, phases []Phase) metadata.Experiment {
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

func (c *toMetadata) addAggressorsToPhase(phaseMetadata metadata.Phase, phaseModel Phase) metadata.Phase {
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

func (c *toMetadata) addMeasurementsToPhases(experimentMetadata metadata.Experiment, measurements []Measurement) metadata.Experiment {
	for _, measurement := range measurements {
		phase := &experimentMetadata.Phases[c.phaseNameToIndex[measurement.PhaseID+experimentMetadata.ID]]
		phase.Measurements = append(phase.Measurements, measurement.Measurement)
	}

	return c.sortMeasurements(experimentMetadata)
}

func (c *toMetadata) sortMeasurements(experimentMetadata metadata.Experiment) metadata.Experiment {
	for _, phase := range experimentMetadata.Phases {
		sort.Sort(phase.Measurements)
	}

	return experimentMetadata
}
