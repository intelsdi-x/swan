package cassandra

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCassandraConnection(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	Convey("While creating Cassandra config with proper parameters", t, func() {
		config, err := cassandra.CreateConfigWithSession("127.0.0.1", "snap")
		Convey("I should receive not nil config", func() {
			So(config, ShouldNotBeNil)
			So(err, ShouldBeNil)
			Convey("Config should have not nil session", func() {
				session := config.CassandraSession()
				So(session, ShouldNotBeNil)
			})
		})
	})
}
