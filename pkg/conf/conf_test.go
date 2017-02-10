package conf

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testAppName       = "testAppName"
	testIPDefaultName = "127.0.0.1"
)

func TestConf(t *testing.T) {
	Convey("While using Conf pkg", t, func() {

		Convey("Log level can be fetched", func() {
			So(LogLevel(), ShouldEqual, logrus.ErrorLevel)
		})

		Convey("Log level can be fetched from env", func() {
			// Default one.
			So(LogLevel(), ShouldEqual, logrus.ErrorLevel)

			os.Setenv(envName(logLevelFlag.Name), "debug")

			ParseFlags()

			// Should be from environment.
			So(LogLevel(), ShouldEqual, logrus.DebugLevel)
		})
	})
}
