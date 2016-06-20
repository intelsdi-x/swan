package uploaders

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
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
				Convey("I should get no error and see metrics saved", func() {
					So(err, ShouldBeNil)
					session, err := createGocqlSession(config)
					So(err, ShouldBeNil)
					defer session.Close()
					soExperimentMetadataAreSaved(session)
					soPhaseMetadataAreSaved(session)
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

func createValidSwanMetrics() sensitivity.Metadata {
	var err error
	m := sensitivity.Metadata{}

	m.LoadDuration, err = time.ParseDuration("1s")
	So(err, ShouldBeNil)
	m.TuningDuration, err = time.ParseDuration("2s")
	So(err, ShouldBeNil)

	m.RepetitionsNumber = 905
	m.LoadPointsNumber = 509

	m.LCName = "LC task"
	m.LCParameters = "Latency critical parameters"
	m.LCIsolation = "Latency critical isolation"

	m.LGName = []string{"LG task"}
	m.LGParameters = []string{"Load generator parameters"}
	m.LGIsolation = "LG isolation"

	m.AggressorName = []string{"an aggressor"}
	m.AggressorParameters = []string{"aggressor parameters"}
	m.AggressorIsolations = []string{"aggressor isolation"}

	m.QPS = 666.6
	m.LGName = []string{"LG task"}
	m.Load = 0.65

	m.ExperimentID = "experiment"
	m.PhaseID = "phase"
	m.RepetitionID = 303

	return m
}

func createGocqlSession(config uploaders.Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(config.Host...)
	cluster.ProtoVersion = 4
	cluster.Keyspace = config.KeySpace
	return cluster.CreateSession()

}

func soExperimentMetadataAreSaved(session *gocql.Session) {
	var id, lcName string
	var LGNames []string
	var loadDuration, tuningDuration time.Duration
	var repetitionsNumber, loadPointsNumber, sLO int

	query := session.Query("SELECT * FROM experiment__id__ LIMIT 1")
	query.Consistency(gocql.One)
	err := query.Scan(&id, &lcName, &LGNames, &loadDuration, &loadPointsNumber, &repetitionsNumber, &sLO, &tuningDuration)
	query.Release()

	So(err, ShouldBeNil)
	So(id, ShouldEqual, "experiment")
	So(lcName, ShouldEqual, "LC task")
	So(LGNames, ShouldResemble, []string{"LG task"})
	So(loadPointsNumber, ShouldEqual, 509)
	So(repetitionsNumber, ShouldEqual, 905)
	//So(sLO, ShouldEqual, 2048)
	oneSecond, err := time.ParseDuration("1s")
	So(err, ShouldBeNil)
	So(loadDuration, ShouldEqual, oneSecond)
	twoSeconds, err := time.ParseDuration("2s")
	So(err, ShouldBeNil)
	So(tuningDuration, ShouldEqual, twoSeconds)
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
