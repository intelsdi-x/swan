package cassandra

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/hailocab/gocassa"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"
)

const (
	experimentTablePrefix  = "experiment"
	phaseTablePrefix       = "phase"
	measurementTablePrefix = "measurement"
)

type repository struct {
	experiment  gocassa.Table
	phase       gocassa.Table
	measurement gocassa.Table
	mapper      *toMetadata
}

// Config stores Cassandra database configuration.
type Config struct {
	Username string
	Password string
	Host     []string
	Port     int
	KeySpace string
}

// NewCassandra created new Cassandra repository.
func NewCassandra(experiment, phase, measurement gocassa.Table) metadata.Repository {
	return &repository{experiment, phase, measurement, &toMetadata{phaseNameToIndex: make(map[string]int)}}
}

// NewKeySpace creates instance of gocassa.KeySpace.
func NewKeySpace(config Config) (gocassa.KeySpace, error) {
	gocql := gocql.NewCluster(config.Host...)
	gocql.ProtoVersion = 4
	gocql.Timeout = 100 * time.Second
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

// NewExperimentTable creates new instance of gocassa.Table representing experiment metadata.
func NewExperimentTable(keySpace gocassa.KeySpace) (gocassa.Table, error) {
	table := keySpace.Table(experimentTablePrefix, &Experiment{}, gocassa.Keys{PartitionKeys: []string{"ID"}})
	err := table.CreateIfNotExist()
	if err != nil {
		return nil, err
	}

	return table, nil
}

// NewPhaseTable creates new instance of gocassa.Table representing phase metadata.
func NewPhaseTable(keySpace gocassa.KeySpace) (gocassa.Table, error) {
	table := keySpace.Table(phaseTablePrefix, &Phase{}, gocassa.Keys{PartitionKeys: []string{"ExperimentID"}, ClusteringColumns: []string{"ID"}})
	err := table.CreateIfNotExist()
	if err != nil {
		return nil, err
	}

	return table, nil
}

// NewMeasurementTable creates new instance of gocassa.Table representing measurement metadata.
func NewMeasurementTable(keySpace gocassa.KeySpace) (gocassa.Table, error) {
	table := keySpace.Table(measurementTablePrefix, &Measurement{}, gocassa.Keys{PartitionKeys: []string{"ExperimentID"}, ClusteringColumns: []string{"PhaseID", "Load"}})
	err := table.CreateIfNotExist()
	if err != nil {
		return nil, err
	}

	return table, nil
}

// Save implements metadata.Repository interface.
func (c repository) Save(metadata metadata.Experiment) error {
	experimentMetrics, phasesMetrics, measurementsMetrics := parseFromMetadata(metadata)
	err := c.experiment.Set(experimentMetrics).Run()
	if err != nil {
		return fmt.Errorf("Experiment metrics saving failed: %s", err.Error())
	}
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

// Fetch implements metadata.Repository interface.
func (c repository) Fetch(experiment string) (metadata.Experiment, error) {
	var experimentModel Experiment
	err := c.experiment.Where(gocassa.Eq("ID", experiment)).ReadOne(&experimentModel).Run()
	if err != nil {
		return metadata.Experiment{}, fmt.Errorf("Experiment metadata fetch failed: %s (experiment: %s)", err.Error(), experiment)
	}

	var phases []Phase
	err = c.phase.Where(gocassa.Eq("ExperimentID", experiment)).Read(&phases).Run()
	if err != nil {
		return metadata.Experiment{}, fmt.Errorf("Phases metadata fetch failed: %s (experiment: %s)", err.Error(), experiment)
	}
	if len(phases) == 0 {
		return metadata.Experiment{}, fmt.Errorf("Phases metadata fetch returned no results (experiment: %s)", experiment)
	}

	var measurements []Measurement
	err = c.measurement.Where(gocassa.Eq("ExperimentID", experiment)).Read(&measurements).Run()
	if err != nil {
		return metadata.Experiment{}, fmt.Errorf("Measurements metadata fetch failed: %s (experiment: %s)", err.Error(), experiment)
	}
	if len(measurements) == 0 {
		return metadata.Experiment{}, fmt.Errorf("Measurements metadata fetch returned no results (experiment: %s)", experiment)
	}

	experimentMetadata := c.mapper.transform(experimentModel, phases, measurements)

	return experimentMetadata, nil
}
