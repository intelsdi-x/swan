package uploaders

import "github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"

func metadataToCassandra(experiment metadata.Experiment) (Experiment, []Phase, []Measurement) {
	experimentMetadata := buildExperimentMetadata(experiment)
	phaseMetadata, measurementMetadata := buildPhaseMetadata(experiment)

	return experimentMetadata, phaseMetadata, measurementMetadata

}

func buildExperimentMetadata(metadata metadata.Experiment) Experiment {
	experimentMetadata := Experiment{}
	experimentMetadata.ID = metadata.ExperimentID
	experimentMetadata.LoadDuration = metadata.LoadDuration
	experimentMetadata.TuningDuration = metadata.TuningDuration
	experimentMetadata.LcName = metadata.LcName
	experimentMetadata.LoadPointsNumber = metadata.LoadPointsNumber
	experimentMetadata.RepetitionsNumber = metadata.RepetitionsNumber
	experimentMetadata.LgNames = append(experimentMetadata.LgNames, metadata.LgNames...)

	return experimentMetadata
}

func buildPhaseMetadata(experiment metadata.Experiment) ([]Phase, []Measurement) {
	var phasesMetadata []Phase
	var measurementsMetadata []Measurement
	for _, metadata := range experiment.Phases {
		phasesMetadata = append(phasesMetadata, Phase{
			ID:                  metadata.ID,
			ExperimentID:        experiment.ExperimentID,
			AggressorNames:      metadata.AggressorNames,
			AggressorIsolations: metadata.AggressorIsolations,
			AggressorParameters: metadata.AggressorParameters,
			LCIsolation:         metadata.LCIsolation,
			LCParameters:        metadata.LCParameters,
		})
		measurementsMetadata = append(measurementsMetadata, buildMeasurementMetadata(metadata, experiment.ExperimentID)...)
	}

	return phasesMetadata, measurementsMetadata
}

func buildMeasurementMetadata(phase metadata.Phase, experiment string) []Measurement {
	var measurementsMetadata []Measurement
	for _, metadata := range phase.Measurements {
		measurementsMetadata = append(measurementsMetadata, Measurement{
			ExperimentID: experiment,
			PhaseID:      phase.ID,
			Load:         metadata.Load,
			LoadPointQPS: metadata.LoadPointQPS,
			LGParameters: metadata.LGParameters,
		})
	}

	return measurementsMetadata
}
