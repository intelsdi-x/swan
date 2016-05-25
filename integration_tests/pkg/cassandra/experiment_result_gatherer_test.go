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

func insertDataIntoCassandra(session *gocql.Session, metrics *cassandra.Metrics) error {
	err := session.Query(`insert into snap.metrics(ns, ver, host, time, boolval, doubleval, labels, strval, tags,
	valtype) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		metrics.Namespace(), metrics.Version(), metrics.Host(), metrics.Time(), metrics.Boolval(),
		metrics.Doubleval(), metrics.Labels(), metrics.Strval(), metrics.Tags(), metrics.Valtype()).Exec()
	if err != nil {
		return err
	}
	return nil
}

func TestValuesGatherer(t *testing.T) {
	// Create fake experiment ID.
	rand.Seed(int64(time.Now().Nanosecond()))
	value := rand.Int()
	experimentID := fmt.Sprintf("%d", value)
	expectedTagsMap := map[string]string{"swan_experiment": experimentID, "swan_phase": "p2", "swan_repetition": "2"}

	//Create Metrics struct that will be inserted into cassandra.
	metrics := cassandra.NewMetrics(experimentID, 1, "abc", time.Now(), false, 10, []string{"b"}, "c", expectedTagsMap, "boolean")

	logrus.SetLevel(logrus.ErrorLevel)
	Convey("While connecting to casandra with proper parameters", t, func() {
		cassandraConfig, err := cassandra.CreateConfigWithSession("127.0.0.1", "snap")
		session := cassandraConfig.CassandraSession()
		Convey("I should receive not empty session", func() {
			So(session, ShouldNotBeNil)
			So(err, ShouldBeNil)
			Convey("I should be able to insert data into cassandra", func() {
				err := insertDataIntoCassandra(session, metrics)
				So(err, ShouldBeNil)
				Convey("and I should be able to receive expected values and close session", func() {
					metricsList := cassandraConfig.GetValuesForGivenExperiment(experimentID)
					resultedMetrics := metricsList[0]

					// Check values of metrics.
					So(resultedMetrics.Namespace(), ShouldEqual, metrics.Namespace())
					So(resultedMetrics.Version(), ShouldEqual, metrics.Version())
					So(resultedMetrics.Host(), ShouldEqual, metrics.Host())
					_, _, resultedDay := resultedMetrics.Time().Date()
					_, _, expectedDay := metrics.Time().Date()
					So(resultedDay, ShouldEqual, expectedDay)
					So(resultedMetrics.Boolval(), ShouldEqual, metrics.Boolval())
					So(resultedMetrics.Doubleval(), ShouldEqual, metrics.Doubleval())
					So(resultedMetrics.Labels()[0], ShouldEqual, metrics.Labels()[0])
					So(resultedMetrics.Strval(), ShouldEqual, metrics.Strval())
					So(resultedMetrics.Tags()["swan_experiment"], ShouldEqual,
						metrics.Tags()["swan_experiment"])
					So(resultedMetrics.Tags()["swan_phase"], ShouldEqual,
						metrics.Tags()["swan_phase"])
					So(resultedMetrics.Tags()["swan_repetition"], ShouldEqual,
						metrics.Tags()["swan_repetition"])
					So(resultedMetrics.Valtype(), ShouldEqual, metrics.Valtype())

					err := cassandraConfig.CloseSession()
					So(err, ShouldBeNil)
				})
			})
		})

	})
}
