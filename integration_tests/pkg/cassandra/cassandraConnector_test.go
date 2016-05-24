package cassandra

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDbConnection(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	Convey("While connecting to casandra with proper parameters", t, func() {
		session, err := cassandra.CreateSession("127.0.0.1", "snap")
		Convey("I should receive not nil session", func() {
			So(session, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}
