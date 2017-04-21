package conf

import (
	"io/ioutil"
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
			level, err := LogLevel()
			So(err, ShouldBeNil)
			So(level, ShouldEqual, logrus.ErrorLevel)
		})

		Convey("Log level can be fetched from env", func() {
			level, err := LogLevel()
			So(err, ShouldBeNil)
			So(level, ShouldEqual, logrus.ErrorLevel)

			os.Setenv(envName(logLevelFlag.Name), "debug")

			ParseFlags()

			// Should be from environment.
			level, err = LogLevel()
			So(err, ShouldBeNil)
			So(level, ShouldEqual, logrus.DebugLevel)
		})

		Convey("Validation for flags from env still works", func() {
			os.Setenv(envName(CassandraConnectionTimeout.Name), "foo-is-not-duration")
			err := ParseFlags()
			So(err, ShouldNotBeNil)
		})

		Convey("Validation for flags loaded from file", func() {
			const testfile = "testfile"
			err := ioutil.WriteFile(testfile, []byte("# comment\nfoo=baz"), os.ModePerm)
			So(err, ShouldBeNil)
			Reset(func() {
				os.Remove(testfile)
			})
			err = LoadConfig(testfile)
			So(err, ShouldBeNil)
			So(os.Getenv("SWAN_foo"), ShouldEqual, "baz")
		})
	})
}
