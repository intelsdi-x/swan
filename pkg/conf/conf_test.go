package conf

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const (
	testAppName       = "testAppName"
	testIPDefaultName = "127.0.0.1"
)

func TestEnvFlag(t *testing.T) {
	Convey("While using Flag struct, it should construct proper swan environment var name", t, func() {
		So(NewStringFlag("test_name", "", "").envName(), ShouldEqual, "SWAN_TEST_NAME")
	})
}

func TestConf(t *testing.T) {
	testReadmePath := path.Join(fs.GetSwanPath(), "pkg", "conf", "test_file.md")
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

		Convey("When some custom String Flag is defined", func() {
			// Register custom flag.
			customFlag := NewStringFlag("custom_string_arg", "help", "default")
			customFlag.clear()
			defer customFlag.clear()

			Convey("Without parse it should be default", func() {
				So(customFlag.Value(), ShouldEqual, "default")
			})

			Convey("When we not defined any environment variable we should have default value after parse", func() {
				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customFlag.defaultValue)
			})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := "customContent"
				os.Setenv(customFlag.envName(), customValue)

				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})

		Convey("When some custom Bool Flag is defined", func() {
			// Register custom flag.
			customFlag := NewBoolFlag("custom_bool_arg", "help", false)
			customFlag.clear()
			defer customFlag.clear()

			Convey("Without parse it should be default", func() {
				So(customFlag.Value(), ShouldEqual, false)
			})

			Convey("When we not defined any environment variable we should have default value after parse", func() {
				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customFlag.defaultValue)
			})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := true
				os.Setenv(customFlag.envName(), fmt.Sprintf("%v", customValue))

				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})

		Convey("When some custom Int Flag is defined", func() {
			// Register custom flag.
			customFlag := NewIntFlag("custom_int_arg", "help", 23424)
			customFlag.clear()
			defer customFlag.clear()

			Convey("Without parse it should be default", func() {
				So(customFlag.Value(), ShouldEqual, 23424)
			})

			Convey("When we not defined any environment variable we should have default value after parse", func() {
				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customFlag.defaultValue)
			})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := 12
				os.Setenv(customFlag.envName(), fmt.Sprintf("%d", customValue))

				err := ParseEnv()
				So(err, ShouldBeNil)
				So(customFlag.Value(), ShouldEqual, customValue)
			})
		})
	})
}
