package cassandra

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"testing"
	"time"
)

func createKeyspace(ip string) (err error) {
	cluster := gocql.NewCluster(ip)
	cluster.ProtoVersion = 4
	cluster.Consistency = gocql.All
	cluster.Timeout = 100 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}

	err = session.Query(
		`CREATE KEYSPACE IF NOT EXISTS snap WITH replication = {
		'class': 'SimpleStrategy','replication_factor':1}`).Exec()
	if err != nil {
		return err
	}
	session.Close()

	return nil
}

func insertDataIntoCassandra(session *gocql.Session, metrics *cassandra.Metrics) (err error) {
	// TODO(CD): Consider getting schema from the cassandra publisher plugin
	err = session.Query(`CREATE TABLE IF NOT EXISTS snap.metrics (
		ns  text,
		ver int,
		host text,
		time timestamp,
		valtype text,
		doubleVal double,
		boolVal boolean,
		strVal text,
		tags map<text,text>,
		PRIMARY KEY ((ns, ver, host), time)
	) WITH CLUSTERING ORDER BY (time DESC);`,
	).Exec()

	if err != nil {
		return err
	}

	err = session.Query(`insert into snap.metrics(
		ns, ver, host, time, boolval,
		doubleval, strval, tags, valtype) values
		(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		metrics.Namespace(), metrics.Version(), metrics.Host(), metrics.Time(), metrics.Boolval(),
		metrics.Doubleval(), metrics.Strval(), metrics.Tags(), metrics.Valtype(),
	).Exec()

	if err != nil {
		return err
	}
	return nil
}

func TestValuesGatherer(t *testing.T) {
	ip := "127.0.0.1"
	Convey("While creating keyspace I should receive no error", t, func() {
		err := createKeyspace(ip)
		So(err, ShouldBeNil)
		// Create fake experiment ID.
		rand.Seed(int64(time.Now().Nanosecond()))
		value := rand.Int()
		experimentID := fmt.Sprintf("%d", value)
		expectedTagsMap := map[string]string{"swan_experiment": experimentID, "swan_phase": "p2", "swan_repetition": "2"}

		//Create Metrics struct that will be inserted into cassandra.
		metrics := cassandra.NewMetrics(experimentID, 1, "abc", time.Now(), false, 10, "c", expectedTagsMap, "boolval")

		logrus.SetLevel(logrus.ErrorLevel)
		Convey("While connecting to Cassandra with proper parameters", func() {
			cassandraConfig, err := cassandra.CreateConfigWithSession(ip, "snap")
			So(err, ShouldBeNil)
			session := cassandraConfig.CassandraSession()
			Convey("I should receive not empty session", func() {
				So(session, ShouldNotBeNil)
				So(err, ShouldBeNil)
				Convey("I should be able to insert data into cassandra", func() {
					err := insertDataIntoCassandra(session, metrics)
					So(err, ShouldBeNil)
					Convey("and I should be able to receive expected values and close session", func() {
						metricsList, err := cassandraConfig.GetValuesForGivenExperiment(experimentID)
						So(len(metricsList), ShouldBeGreaterThan, 0)
						So(err, ShouldBeNil)
						resultedMetrics := metricsList[0]

						// Check values of metrics.
						So(resultedMetrics.Namespace(), ShouldEqual, metrics.Namespace())
						So(resultedMetrics.Version(), ShouldEqual, metrics.Version())
						So(resultedMetrics.Host(), ShouldEqual, metrics.Host())

						// Cassandra stores time values in UTC by default. So, we
						// convert the expected time value to UTC to avoid discrepancies
						// in the interpreted calendar date and the test flakiness
						// that could cause. For completeness, we also pre-emptively
						// convert the result time to UTC in case the database is
						// configured to use a non-default TZ.
						_, _, resultedDay := resultedMetrics.Time().UTC().Date()
						_, _, expectedDay := metrics.Time().UTC().Date()

						So(resultedDay, ShouldEqual, expectedDay)
						So(resultedMetrics.Boolval(), ShouldEqual, metrics.Boolval())
						So(resultedMetrics.Doubleval(), ShouldEqual, metrics.Doubleval())
						So(resultedMetrics.Strval(), ShouldEqual, metrics.Strval())
						So(resultedMetrics.Tags()["swan_experiment"], ShouldEqual,
							metrics.Tags()["swan_experiment"])
						So(resultedMetrics.Tags()["swan_phase"], ShouldEqual,
							metrics.Tags()["swan_phase"])
						So(resultedMetrics.Tags()["swan_repetition"], ShouldEqual,
							metrics.Tags()["swan_repetition"])
						So(resultedMetrics.Valtype(), ShouldEqual, metrics.Valtype())

						err = cassandraConfig.CloseSession()
						So(err, ShouldBeNil)
					})
				})
			})

		})
	})
}
