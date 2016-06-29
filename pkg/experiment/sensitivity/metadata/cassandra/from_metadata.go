package cassandra

import "github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"

func metadataToCassandra(experiment metadata.Experiment) (Experiment, []Phase, []Measurement) {
	experimentMetadata := buildExperimentMetadata(experiment)
	phaseMetadata, measurementMetadata := buildPhaseMetadata(experiment)

	return experimentMetadata, phaseMetadata, measurementMetadata

}

func buildExperimentMetadata(metadata metadata.Experiment) Experiment {
	return Experiment{
		BaseExperiment: metadata.BaseExperiment,
	}
}

func buildPhaseMetadata(experiment metadata.Experiment) ([]Phase, []Measurement) {
	var phasesMetadata []Phase
	var measurementsMetadata []Measurement
	for _, metadata := range experiment.Phases {
		var names, isolations, parameters []string
		for _, aggressor := range metadata.Aggressors {
			names = append(names, aggressor.Name)
			isolations = append(isolations, aggressor.Isolation)
			parameters = append(parameters, aggressor.Parameters)
		}
		phasesMetadata = append(phasesMetadata, Phase{
			BasePhase:           metadata.BasePhase,
			ExperimentID:        experiment.ID,
			AggressorNames:      names,
			AggressorIsolations: isolations,
			AggressorParameters: parameters,
		})
		measurementsMetadata = append(measurementsMetadata, buildMeasurementMetadata(metadata, experiment.ID)...)
	}

	return phasesMetadata, measurementsMetadata
}

func buildMeasurementMetadata(phase metadata.Phase, experiment string) []Measurement {
	var measurementsMetadata []Measurement
	for _, metadata := range phase.Measurements {
		measurementsMetadata = append(measurementsMetadata, Measurement{
			ExperimentID: experiment,
			PhaseID:      phase.ID,
			Measurement:  metadata,
		})
	}

	return measurementsMetadata
}
