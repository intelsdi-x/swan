package uploaders

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/hailocab/gocassa"
	"github.com/intelsdi-x/swan/pkg/metrics"
)

const experimentTablePrefix = "experiment"
const phaseTablePrefix = "phase"
const measurementTablePrefix = "measurement"

type cassandra struct {
	experiment  gocassa.Table
	phase       gocassa.Table
	measurement gocassa.Table
}

type experiment struct {
	ID               string
	TestingDuration  time.Duration
	LcName           string
	LcParameters     string
	LcIsolation      string
	LgName           string
	LgParameters     string
	LgIsolation      string
	Repetitions      int
	LoadPointsNumber int
	SLO              int
}

type phase struct {
	ID            string
	ExperimentID  string
	AggressorName string
}

type measurement struct {
	ID                  int
	PhaseID             string
	ExperimentID        string
	AggressorParameters string
	HandledQPS          int
	TargetQPS           int
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
func NewCassandra(config Config) (metrics.Uploader, error) {
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
	measurementTable := keySpace.Table(measurementTablePrefix, &measurement{}, gocassa.Keys{PartitionKeys: []string{"ID", "ExperimentID", "PhaseID"}})
	experimentTable.CreateIfNotExist()
	phaseTable.CreateIfNotExist()
	measurementTable.CreateIfNotExist()

	return &cassandra{experimentTable, phaseTable, measurementTable}, nil
}

//SendMetrics implements metrics.Uploader interface
func (c cassandra) SendMetrics(metrics metrics.Swan) error {
	experimentMetrics := c.buildExperimentMetadata(metrics)
	err := c.experiment.Set(experimentMetrics).Run()
	if err != nil {
		return fmt.Errorf("Experiment metrics saving failed: %s", err.Error())
	}
	phaseMetrics := c.buildPhaseMetadata(metrics)
	err = c.phase.Set(phaseMetrics).Run()
	if err != nil {
		return fmt.Errorf("Phase metrics saving failed: %s", err.Error())
	}
	measurementMetrics := c.buildMeasurementMetadata(metrics)
	err = c.measurement.Set(measurementMetrics).Run()
	if err != nil {
		return fmt.Errorf("Measurement metrics saving failed: %s", err.Error())
	}

	return nil

}

func (c cassandra) buildExperimentMetadata(metrics metrics.Swan) experiment {
	experimentMetrics := experiment{}
	experimentMetrics.ID = metrics.Tags.ExperimentID
	experimentMetrics.TestingDuration = metrics.Metrics.LoadDuration
	experimentMetrics.LcName = metrics.Metrics.LCName
	experimentMetrics.LcParameters = metrics.Metrics.LCParameters
	experimentMetrics.LcIsolation = metrics.Metrics.LCIsolation
	experimentMetrics.LgParameters = metrics.Metrics.LGParameters
	experimentMetrics.LgName = metrics.Metrics.LGName
	experimentMetrics.LgIsolation = metrics.Metrics.LGIsolation
	experimentMetrics.LoadPointsNumber = metrics.Metrics.LoadPointsNumber
	experimentMetrics.Repetitions = metrics.Metrics.RepetitionsNumber
	experimentMetrics.SLO = metrics.Metrics.SLO

	return experimentMetrics
}

func (c cassandra) buildPhaseMetadata(metrics metrics.Swan) phase {
	phaseMetrics := phase{}
	phaseMetrics.ID = metrics.Tags.PhaseID
	phaseMetrics.ExperimentID = metrics.Tags.ExperimentID
	phaseMetrics.AggressorName = metrics.Metrics.AggressorName

	return phaseMetrics
}

func (c cassandra) buildMeasurementMetadata(metrics metrics.Swan) measurement {
	measurementMetrics := measurement{}
	measurementMetrics.ID = metrics.Tags.RepetitionID
	measurementMetrics.ExperimentID = metrics.Tags.ExperimentID
	measurementMetrics.PhaseID = metrics.Tags.PhaseID
	measurementMetrics.AggressorParameters = metrics.Metrics.AggressorParameters

	return measurementMetrics
}
