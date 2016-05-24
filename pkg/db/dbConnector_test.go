package db

import (
	"github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDbConnection(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	Convey("While connecting to casandra with proper parameters", t, func() {
		cluster := configureCluster("127.0.0.1", "snap")
		So(cluster, ShouldNotBeNil)
		session, err := createSession(cluster)
		Convey("I should receive not nil session", func() {
			So(session, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}
