package uploaders

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/hailocab/gocassa"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"
)

const (
	experimentTablePrefix  = "experiment"
	phaseTablePrefix       = "phase"
	measurementTablePrefix = "measurement"
)

type cassandra struct {
	experiment  gocassa.Table
	phase       gocassa.Table
	measurement gocassa.Table
}

// Experiment is a gocassa model for experiment metadata
type Experiment struct {
	ID                string
	LoadDuration      time.Duration
	TuningDuration    time.Duration
	LcName            string
	LgNames           []string
	RepetitionsNumber int
	LoadPointsNumber  int
	SLO               int
}

// Phase is a gocassa model for phase metadata
type Phase struct {
	ID                  string
	ExperimentID        string
	LCParameters        string
	LCIsolation         string
	AggressorNames      []string
	AggressorParameters []string
	AggressorIsolations []string
	Load                float64
	LoadPointQPS        float64
}

// Measurement is a gocassa model for phase metadata
type Measurement struct {
	PhaseID      string
	ExperimentID string
	Load         float64
	LoadPointQPS float64
	LGParameters []string
}

// Config stores Cassandra database configuration
type Config struct {
	Username string
	Password string
	Host     []string
	Port     int
	KeySpace string
}

// NewCassandra created new Cassandra Uploader
func NewCassandra(experiment, phase, measurement gocassa.Table) sensitivity.Uploader {
	return &cassandra{experiment, phase, measurement}
}

func NewKeySpace(config Config) (gocassa.KeySpace, error) {
	gocql := gocql.NewCluster(config.Host...)
	gocql.ProtoVersion = 4
	session, err := gocql.CreateSession()
	if err != nil {
		err = fmt.Errorf("Creating gocql session failed: %s", err.Error())
		return nil, err
	}
	executor := gocassa.GoCQLSessionToQueryExecutor(session)
	conn := gocassa.NewConnection(executor)
	err = conn.CreateKeySpace(config.KeySpace)
	if err != nil {
		err = fmt.Errorf("Creating keyspace failed: %s", err.Error())
		return nil, err
	}
	keySpace := conn.KeySpace(config.KeySpace)

	return keySpace, nil
}

func NewExperimentTable(keySpace gocassa.KeySpace) gocassa.Table {
	table := keySpace.Table(experimentTablePrefix, &Experiment{}, gocassa.Keys{PartitionKeys: []string{"ID"}})
	table.CreateIfNotExist()

	return table
}

func NewPhaseTable(keySpace gocassa.KeySpace) gocassa.Table {
	table := keySpace.Table(phaseTablePrefix, &Phase{}, gocassa.Keys{PartitionKeys: []string{"ExperimentID"}, ClusteringColumns: []string{"ID"}})
	table.CreateIfNotExist()

	return table
}

func NewMeasurementTable(keySpace gocassa.KeySpace) gocassa.Table {
	table := keySpace.Table(measurementTablePrefix, &Measurement{}, gocassa.Keys{PartitionKeys: []string{"ExperimentID"}, ClusteringColumns: []string{"PhaseID", "Load"}})
	table.CreateIfNotExist()

	return table
}

//SendMetrics implements metrics.Uploader interface
func (c cassandra) SendMetadata(metadata metadata.Experiment) error {
	experimentMetrics := c.buildExperimentMetadata(metadata)
	err := c.experiment.Set(experimentMetrics).Run()
	if err != nil {
		return fmt.Errorf("Experiment metrics saving failed: %s", err.Error())
	}
	phasesMetrics, measurementsMetrics := c.buildPhaseMetadata(metadata)
	for _, phase := range phasesMetrics {
		err = c.phase.Set(phase).Run()
		if err != nil {
			return fmt.Errorf("Phase metrics saving failed: %s (experiment: %s, phase: %s)", err.Error(), metadata.ID, phase.ID)
		}
	}
	for _, measurement := range measurementsMetrics {
		err = c.measurement.Set(measurement).Run()
	}

	return nil

}

func (c cassandra) GetMetadata(experiment string) (metadata.Experiment, error) {
	var experimentModel Experiment
	err := c.experiment.Where(gocassa.Eq("ID", experiment)).ReadOne(&experimentModel).Run()
	if err != nil {
		return metadata.Experiment{}, fmt.Errorf("Experiment metadata fetch failed: %s (experiment: %s)", err.Error(), experiment)
	}
	experimentMetadata := metadata.Experiment{
		ID:                experimentModel.ID,
		LoadDuration:      experimentModel.LoadDuration,
		TuningDuration:    experimentModel.TuningDuration,
		LcName:            experimentModel.LcName,
		LgNames:           experimentModel.LgNames,
		RepetitionsNumber: experimentModel.RepetitionsNumber,
		LoadPointsNumber:  experimentModel.LoadPointsNumber,
		SLO:               experimentModel.SLO,
	}
	var phases []Phase
	err = c.phase.Where(gocassa.Eq("ExperimentID", experiment)).Read(&phases).Run()
	if err != nil {
		return metadata.Experiment{}, fmt.Errorf("Phases metadata fetch failed: %s (experiment: %s)", err.Error(), experiment)
	}
	if len(phases) == 0 {
		return metadata.Experiment{}, fmt.Errorf("Phases metadata fetch returned no results (experiment: %s)", experiment)
	}
	for _, phase := range phases {
		phaseMetadata := metadata.Phase{
			ID:                  phase.ID,
			LCParameters:        phase.LCParameters,
			LCIsolation:         phase.LCIsolation,
			AggressorNames:      phase.AggressorNames,
			AggressorIsolations: phase.AggressorIsolations,
			AggressorParameters: phase.AggressorParameters,
		}
		var measurements []Measurement
		err = c.measurement.Where(gocassa.Eq("ExperimentID", experiment), gocassa.Eq("PhaseID", phase.ID)).Read(&measurements).Run()
		if err != nil {
			return metadata.Experiment{}, fmt.Errorf("Measurement metadata fetch failed: %s (experiment: %s, phase: %s)", err.Error(), experiment, phase.ID)
		}
		if len(measurements) == 0 {
			return metadata.Experiment{}, fmt.Errorf("Measurement metadata fetch returned no results (experiment: %s, phase: %s)", experiment, phase.ID)
		}
		for _, measurement := range measurements {
			phaseMetadata.Measurements = append(phaseMetadata.Measurements, metadata.Measurement{
				Load:         measurement.Load,
				LoadPointQPS: measurement.LoadPointQPS,
				LGParameters: measurement.LGParameters,
			})
		}
		experimentMetadata.Phases = append(experimentMetadata.Phases, phaseMetadata)
	}

	return experimentMetadata, nil
}

func (c cassandra) buildExperimentMetadata(metadata metadata.Experiment) Experiment {
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

func (c cassandra) buildPhaseMetadata(experiment metadata.Experiment) ([]Phase, []Measurement) {
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
		measurementsMetadata = append(measurementsMetadata, c.buildMeasurementMetadata(metadata, experiment.ExperimentID)...)
	}

	return phasesMetadata, measurementsMetadata
}

func (c cassandra) buildMeasurementMetadata(phase metadata.Phase, experiment string) []Measurement {
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
