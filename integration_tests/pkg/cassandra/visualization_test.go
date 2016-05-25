package cassandra

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDrawTable(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	Convey("While drawing", t, func() {
		cassandra.DrawTable("a")
	})
}
