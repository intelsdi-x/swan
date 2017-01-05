package conf

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testAppName       = "testAppName"
	testIPDefaultName = "127.0.0.1"
)

func TestConf(t *testing.T) {
	testReadmePath := path.Join(fs.GetAthenaPath(), "pkg", "conf", "test_file.md")
	Convey("While using Conf pkg", t, func() {
		logLevelFlag.clear()
		defer logLevelFlag.clear()

		SetAppName(testAppName)
		SetHelpPath(testReadmePath)

		Convey("Name and help should match to specified one", func() {
			So(AppName(), ShouldEqual, testAppName)

			readmeData, err := ioutil.ReadFile(testReadmePath)
			if err != nil {
				panic(err)
			}
			So(app.Help, ShouldEqual, string(readmeData)[:])
		})

		Convey("Log level can be fetched", func() {
			So(LogLevel(), ShouldEqual, logrus.ErrorLevel)
		})

		Convey("Log level can be fetched from env", func() {
			// Default one.
			So(LogLevel(), ShouldEqual, logrus.ErrorLevel)

			os.Setenv(logLevelFlag.envName(), "debug")

			err := ParseEnv()
			So(err, ShouldBeNil)

			// Should be from environment.
			So(LogLevel(), ShouldEqual, logrus.DebugLevel)
		})
	})
}
