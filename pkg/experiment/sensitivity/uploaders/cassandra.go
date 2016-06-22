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
	LGParameters        []string
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
func NewCassandra(config Config) (sensitivity.Uploader, error) {
	gocql := gocql.NewCluster(config.Host...)
	gocql.ProtoVersion = 4
	session, err := gocql.CreateSession()
	if err != nil {
		err = fmt.Errorf("Creating gocql session failed: %s", err.Error())
		return nil, err
	}
	executor := gocassa.GoCQLSessionToQueryExecutor(session)
	conn := gocassa.NewConnection(executor)
	conn.CreateKeySpace(config.KeySpace)
	keySpace := conn.KeySpace(config.KeySpace)

	experimentTable := keySpace.Table(experimentTablePrefix, &Experiment{}, gocassa.Keys{PartitionKeys: []string{"ID"}})
	phaseTable := keySpace.Table(phaseTablePrefix, &Phase{}, gocassa.Keys{PartitionKeys: []string{"ID", "ExperimentID"}})
	measurementTable := keySpace.Table(measurementTablePrefix, &Measurement{}, gocassa.Keys{PartitionKeys: []string{"ExperimentID"}, ClusteringColumns: []string{"PhaseID", "Load"}})
	experimentTable.CreateIfNotExist()
	phaseTable.CreateIfNotExist()
	measurementTable.CreateIfNotExist()

	return &cassandra{experimentTable, phaseTable, measurementTable}, nil
}

//SendMetrics implements metrics.Uploader interface
func (c cassandra) SendMetadata(metadata metadata.Experiment) error {
	experimentMetrics := c.buildExperimentMetadata(metadata)
	err := c.experiment.Set(experimentMetrics).Run()
	if err != nil {
		return fmt.Errorf("Experiment metrics saving failed: %s", err.Error())
	}
	phasesMetrics := c.buildPhaseMetadata(metadata)
	for _, phase := range phasesMetrics {
		err = c.phase.Set(phase).Run()
		if err != nil {
			return fmt.Errorf("Phase metrics saving failed: %s (experiment: %s, phase: %s)", err.Error(), metadata.ID, phase.ID)
		}
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

func (c cassandra) buildPhaseMetadata(experiment metadata.Experiment) []Phase {
	var phasesMetadata []Phase
	for _, metadata := range experiment.Phases {
		phaseMetadata := Phase{}
		phaseMetadata.ID = metadata.ID
		phaseMetadata.ExperimentID = experiment.ID
		phaseMetadata.AggressorNames = metadata.AggressorNames
		phaseMetadata.AggressorIsolations = metadata.AggressorIsolations
		phaseMetadata.AggressorParameters = metadata.AggressorParameters
		phaseMetadata.LCIsolation = metadata.LCIsolation
		phaseMetadata.LCParameters = metadata.LCParameters
		phaseMetadata.LGParameters = metadata.LGParameters
		phaseMetadata.Load = metadata.Load
		phaseMetadata.LoadPointQPS = metadata.LoadPointQPS
		phasesMetadata = append(phasesMetadata, phaseMetadata)
	}

	return phasesMetadata
}
