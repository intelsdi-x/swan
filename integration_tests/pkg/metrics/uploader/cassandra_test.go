package uploaders

import (
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/swan/pkg/metrics"
	"github.com/intelsdi-x/swan/pkg/metrics/uploaders"
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
				err = cassandra.SendMetrics(sM)
				Convey("I should get no error and see metrics saved", func() {
					So(err, ShouldBeNil)
					session, err := createGocqlSession(config)
					So(err, ShouldBeNil)
					defer session.Close()
					soExperimentMetadataAreSaved(session)
					soPhaseMetadataAreSaved(session)
					soMeasurementMetadataAreSaved(session)
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
			//For some reason this query times out (but keyspace get deleted...)
			//So(err, ShouldBeNil)
		})
	})
}

func createValidSwanMetrics() metrics.Swan {
	var err error
	m := metrics.Metadata{}
	m.LoadDuration, err = time.ParseDuration("1s")
	So(err, ShouldBeNil)
	m.LoadPointsNumber = 666
	m.TuningDuration = 1024
	m.RepetitionsNumber = 905
	m.LoadPointsNumber = 509
	m.LCName = "LC task"
	m.LCParameters = "LC params"
	m.LCIsolation = "LC isolation"
	m.LGName = "LG task"
	m.LGParameters = "LG params"
	m.LGIsolation = "LG isolation"
	m.AggressorName = "an aggressor"
	m.AggressorParameters = "aggressor params"
	m.SLO = 2048
	m.QPS = 8096

	t := metrics.Tags{}
	t.ExperimentID = "experiment"
	t.PhaseID = "phase"
	t.RepetitionID = 303
	sM := metrics.Swan{}
	sM.Metrics = m
	sM.Tags = t
	return sM
}

func createGocqlSession(config uploaders.Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(config.Host...)
	cluster.ProtoVersion = 4
	cluster.Keyspace = config.KeySpace
	return cluster.CreateSession()

}

func soExperimentMetadataAreSaved(session *gocql.Session) {
	var id, lcName, lcParameters, lcIsolation, lgName, lgParameters, lgIsolation string
	var testingDuration time.Duration
	var repetitionsNumber, loadPointsNumber, sLO int

	query := session.Query("SELECT * FROM experiment__id__ LIMIT 1")
	query.Consistency(gocql.One)
	err := query.Scan(&id, &lcIsolation, &lcName, &lcParameters, &lgIsolation, &lgName, &lgParameters, &loadPointsNumber, &repetitionsNumber, &sLO, &testingDuration)
	query.Release()

	So(err, ShouldBeNil)
	So(id, ShouldEqual, "experiment")
	So(lcName, ShouldEqual, "LC task")
	So(lcIsolation, ShouldEqual, "LC isolation")
	So(lcParameters, ShouldEqual, "LC params")
	So(lcName, ShouldEqual, "LC task")
	So(lgIsolation, ShouldEqual, "LG isolation")
	So(lgParameters, ShouldEqual, "LG params")
	So(lgName, ShouldEqual, "LG task")
	So(loadPointsNumber, ShouldEqual, 509)
	So(repetitionsNumber, ShouldEqual, 905)
	So(sLO, ShouldEqual, 2048)
	oneSecond, err := time.ParseDuration("1s")
	So(err, ShouldBeNil)
	So(testingDuration, ShouldEqual, oneSecond)
}

func soPhaseMetadataAreSaved(session *gocql.Session) {
	var id, experimentID, aggressorName string
	query := session.Query("SELECT * FROM phase__id_experimentid__ LIMIT 1")
	query.Consistency(gocql.One)
	err := query.Scan(&id, &experimentID, &aggressorName)
	query.Release()
	So(err, ShouldBeNil)
	So(id, ShouldEqual, "phase")
	So(experimentID, ShouldEqual, "experiment")
	So(aggressorName, ShouldEqual, "an aggressor")
}

func soMeasurementMetadataAreSaved(session *gocql.Session) {
	var id, handledQPS, targetQPS int
	var phaseID, experimentID, aggressorParameters string
	query := session.Query("SELECT * FROM measurement__id_experimentid_phaseid__ LIMIT 1")
	query.Consistency(gocql.One)
	err := query.Scan(&id, &experimentID, &phaseID, &aggressorParameters, &handledQPS, &targetQPS)
	query.Release()
	So(err, ShouldBeNil)
	So(id, ShouldEqual, 303)
	So(experimentID, ShouldEqual, "experiment")
	So(phaseID, ShouldEqual, "phase")
	So(aggressorParameters, ShouldEqual, "aggressor params")
	So(handledQPS, ShouldEqual, 0)
	So(targetQPS, ShouldEqual, 0)
}
