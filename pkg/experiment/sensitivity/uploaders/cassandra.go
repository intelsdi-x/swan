package uploaders

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/hailocab/gocassa"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
)

const experimentTablePrefix = "experiment"
const phaseTablePrefix = "phase"
const measurementTablePrefix = "measurement"

type cassandra struct {
	experiment gocassa.Table
	phase      gocassa.Table
}

type experiment struct {
	ID                string
	LoadDuration      time.Duration
	TuningDuration    time.Duration
	LcName            string
	LgNames           []string
	RepetitionsNumber int
	LoadPointsNumber  int
	SLO               int
}

type phase struct {
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

	experimentTable := keySpace.Table(experimentTablePrefix, &experiment{}, gocassa.Keys{PartitionKeys: []string{"ID"}})
	phaseTable := keySpace.Table(phaseTablePrefix, &phase{}, gocassa.Keys{PartitionKeys: []string{"ID", "ExperimentID"}})
	experimentTable.CreateIfNotExist()
	phaseTable.CreateIfNotExist()

	return &cassandra{experimentTable, phaseTable}, nil
}

//SendMetrics implements metrics.Uploader interface
func (c cassandra) SendMetadata(metadata sensitivity.Metadata) error {
	experimentMetrics := c.buildExperimentMetadata(metadata)
	err := c.experiment.Set(experimentMetrics).Run()
	if err != nil {
		return fmt.Errorf("Experiment metrics saving failed: %s", err.Error())
	}
	phaseMetrics := c.buildPhaseMetadata(metadata)
	err = c.phase.Set(phaseMetrics).Run()
	if err != nil {
		return fmt.Errorf("Phase metrics saving failed: %s", err.Error())
	}

	return nil

}

func (c cassandra) buildExperimentMetadata(metadata sensitivity.Metadata) experiment {
	experimentMetadata := experiment{}
	experimentMetadata.ID = metadata.ExperimentID
	experimentMetadata.LoadDuration = metadata.LoadDuration
	experimentMetadata.TuningDuration = metadata.TuningDuration
	experimentMetadata.LcName = metadata.LCName
	experimentMetadata.LoadPointsNumber = metadata.LoadPointsNumber
	experimentMetadata.RepetitionsNumber = metadata.RepetitionsNumber
	experimentMetadata.LgNames = append(experimentMetadata.LgNames, metadata.LGName...)

	return experimentMetadata
}

func (c cassandra) buildPhaseMetadata(metadata sensitivity.Metadata) phase {
	phaseMetadata := phase{}
	phaseMetadata.ID = metadata.PhaseID
	phaseMetadata.ExperimentID = metadata.ExperimentID
	phaseMetadata.AggressorNames = append(phaseMetadata.AggressorNames, metadata.AggressorName...)
	phaseMetadata.AggressorIsolations = metadata.AggressorIsolations
	phaseMetadata.AggressorParameters = metadata.AggressorParameters
	phaseMetadata.LCIsolation = metadata.LCIsolation
	phaseMetadata.LCParameters = metadata.LCParameters
	phaseMetadata.LGParameters = metadata.LGParameters
	phaseMetadata.Load = metadata.Load
	phaseMetadata.LoadPointQPS = metadata.QPS

	return phaseMetadata
}
