package db

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"testing"
)

func insertDataIntoCassandra(session *gocql.Session, tagsMap map[string]string) error {
	err := session.Query(`insert into snap.metrics(ns, ver, host, time, boolval, doubleval, labels, strval, tags,
	valtype) values (?, 1, 'abc', '2013-05-13 09:42:51', False, 10, ['b'], 'c', ?, 'boolean')`,
		tagsMap["swan_experiment"], tagsMap).Exec()
	if err != nil {
		return err
	}
	return nil
}

func TestValuesGatherer(t *testing.T) {
	r := rand.New(rand.NewSource(99))
	experimentName := fmt.Sprintf("fakeExperimentName%f", r.Float32())
	expectedValue := 10
	expectedTagsMap := map[string]string{"swan_experiment": experimentName, "swan_phase": "p2", "swan_repetition": "2"}
	logrus.SetLevel(logrus.ErrorLevel)
	Convey("While connecting to casandra with proper parameters", t, func() {
		cluster := configureCluster("127.0.0.1", "snap")
		So(cluster, ShouldNotBeNil)
		session, err := createSession(cluster)
		Convey("I should receive not empty session", func() {
			So(session, ShouldNotBeNil)
			So(err, ShouldBeNil)
			err := insertDataIntoCassandra(session, expectedTagsMap)
			So(err, ShouldBeNil)
			Convey("and I should be able to receive expected values and tags", func() {
				valuesList, tagsList := getValuesAndTagsForGivenExperiment(session, experimentName)
				So(valuesList[0], ShouldEqual, expectedValue)
				So(tagsList[0]["swan_experiment"], ShouldEqual, expectedTagsMap["swan_experiment"])
			})
		})

	})
}
