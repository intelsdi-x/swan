package uploaders

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/uploaders"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCassandraUploading(t *testing.T) {
	Convey("When I connect to Cassandra instance with incorrect configuration", t, func() {
		config := uploaders.Config{}
		cassandra, err := uploaders.NewCassandra(config)
		Convey("I should get an error and no Uploader instance", func() {
			So(err, ShouldNotBeNil)
			So(cassandra, ShouldBeNil)
		})
	})

	Convey("When I connect to Cassandra instance with correct configuration", t, func() {
		config := uploaders.Config{}
		config.Host = []string{"127.0.0.1"}
		config.KeySpace = "testing_keyspace"
		cassandra, err := uploaders.NewCassandra(config)
		Convey("I should get no error and Uploader instance", func() {
			So(err, ShouldBeNil)
			So(cassandra, ShouldNotBeNil)
			Convey("When I pass SwanMetrics instance", func() {
				sM := createValidSwanMetrics()
				err = cassandra.SendMetadata(sM)
				So(err, ShouldBeNil)
				Convey("I should get no error and see metrics saved", func() {
					metadata, err := cassandra.GetMetadata("experiment")
					So(err, ShouldBeNil)
					soExperimentMetadataAreSaved(metadata)
					/*session, err := createGocqlSession(config)
					So(err, ShouldBeNil)
					defer session.Close()
					soExperimentMetadataAreSaved(session)
					soPhaseMetadataAreSaved(session)*/
				})
				Convey("I should get an error when I try to get metadata for non existing experiment", func() {
					metadata, err := cassandra.GetMetadata("non existing experiment")
					So(err, ShouldNotBeNil)
					So(metadata.ID, ShouldEqual, "")
				})
			})
		})

		Reset(func() {
			session, err := createGocqlSession(config)
			So(err, ShouldBeNil)
			defer session.Close()
			query := session.Query(fmt.Sprintf("DROP KEYSPACE %s", config.KeySpace))
			defer query.Release()
			err = query.Exec()
			//For some reason this query times out (but keyspace gets deleted...)
			//So(err, ShouldBeNil)
		})
	})
}

func createValidSwanMetrics() metadata.Experiment {
	meta := metadata.Experiment{}
	meta.RepetitionsNumber = 905
	meta.LoadPointsNumber = 509

	meta.LcName = "Latency critical task"
	//m.LCParameters = "Latency critical parameters"
	//m.LCIsolation = "Latency critical isolation"

	meta.LgNames = []string{"Load generator task"}
	//m.LGParameters = []string{"Load generator parameters"}
	//m.LGIsolation = "LG isolation"

	//m.AggressorName = []string{"an aggressor"}
	//m.AggressorParameters = []string{"aggressor parameters"}
	//m.AggressorIsolations = []string{"aggressor isolation"}

	//m.QPS = 666.6
	//m.Load = 0.65

	meta.ExperimentID = "experiment"
	//m.PhaseID = "phase
	//m.RepetitionID = 303

	meta.LoadDuration, _ = time.ParseDuration("1s")
	meta.TuningDuration, _ = time.ParseDuration("2s")
	meta.AddPhase(metadata.Phase{
		ID:                  "phase",
		LCParameters:        "Latency critival parameters",
		LCIsolation:         "Latency critical isolation",
		AggressorNames:      []string{"an aggressor"},
		AggressorParameters: []string{"aggressor parameters"},
		AggressorIsolations: []string{"aggressor isolation"},
	})
	meta.Phases[0].AddMeasurement(metadata.Measurement{
		Load:         0.65,
		LoadPointQPS: 666.6,
		LGParameters: []string{"Load generator parameters"},
	})

	return meta
}

func createGocqlSession(config uploaders.Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(config.Host...)
	cluster.ProtoVersion = 4
	cluster.Keyspace = config.KeySpace
	return cluster.CreateSession()

}

func soExperimentMetadataAreSaved(experiment metadata.Experiment) {
	/*var id, lcName string
	var LGNames []string
	var loadDuration, tuningDuration time.Duration
	var repetitionsNumber, loadPointsNumber, sLO int

	query := session.Query("SELECT * FROM experiment__id__ LIMIT 1")
	query.Consistency(gocql.One)
	err := query.Scan(&id, &lcName, &LGNames, &loadDuration, &loadPointsNumber, &repetitionsNumber, &sLO, &tuningDuration)
	query.Release()

	So(err, ShouldBeNil)*/
	So(experiment.ID, ShouldEqual, "experiment")
	So(experiment.LcName, ShouldEqual, "Latency critical task")
	So(experiment.LgNames, ShouldResemble, []string{"Load generator task"})
	So(experiment.LoadPointsNumber, ShouldEqual, 509)
	So(experiment.RepetitionsNumber, ShouldEqual, 905)

	oneSecond, err := time.ParseDuration("1s")
	So(err, ShouldBeNil)
	So(experiment.LoadDuration, ShouldEqual, oneSecond)
	twoSeconds, err := time.ParseDuration("2s")
	So(err, ShouldBeNil)
	So(experiment.TuningDuration, ShouldEqual, twoSeconds)
}

func soPhaseMetadataAreSaved(session *gocql.Session) {
	var id, experimentID, lcIsolation, lcParameters string
	var aggressorsIsolations, aggressorName, aggressorsParameters, lgParameters []string
	var load, loadPointsQPS float64

	query := session.Query("SELECT * FROM phase__id_experimentid__ LIMIT 1")
	query.Consistency(gocql.One)
	err := query.Scan(&id, &experimentID, &aggressorsIsolations, &aggressorName, &aggressorsParameters, &lcIsolation, &lcParameters, &lgParameters, &load, &loadPointsQPS)
	query.Release()

	So(err, ShouldBeNil)
	So(id, ShouldEqual, "phase")
	So(experimentID, ShouldEqual, "experiment")

	So(aggressorsIsolations, ShouldResemble, []string{"aggressor isolation"})
	So(aggressorName, ShouldResemble, []string{"an aggressor"})
	So(aggressorsParameters, ShouldResemble, []string{"aggressor parameters"})
	So(lcIsolation, ShouldEqual, "Latency critical isolation")
	So(lcParameters, ShouldEqual, "Latency critical parameters")
	So(lgParameters, ShouldResemble, []string{"Load generator parameters"})

	So(load, ShouldEqual, 0.65)
	So(loadPointsQPS, ShouldEqual, 666.6)
}
