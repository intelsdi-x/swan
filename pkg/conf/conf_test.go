package conf

import (
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
	customArgKey      = "custom_arg"
	customArgHelp     = "help"
	customArgDefault  = "default"
)

func clearEnv() {
	// Clear all environment variables in context of that test.
	os.Unsetenv(changeToEnvName(logLevelArg))
	os.Unsetenv(changeToEnvName(ipAddressArg))
	os.Unsetenv(changeToEnvName(customArgKey))
}

func TestChangeToEnvName(t *testing.T) {
	Convey("While using changeToEnvName function", t, func() {
		name := "test_name"
		envName := "SWAN_TEST_NAME"
		So(changeToEnvName(name), ShouldEqual, envName)
	})
}

func TestConf(t *testing.T) {
	testReadmePath := path.Join(fs.GetSwanPath(), "pkg", "conf", "test_file.md")
	Convey("While using Config", t, func() {
		clearEnv()
		defer clearEnv()

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

		Convey("Default IP and log level can be fetched", func() {
			So(LogLevel(), ShouldEqual, logrus.ErrorLevel)
			So(IPAddress(), ShouldEqual, testIPDefaultName)
		})

		Convey("Custom IP and log level can be fetched from env", func() {
			// Default one.
			So(LogLevel(), ShouldEqual, logrus.ErrorLevel)
			So(IPAddress(), ShouldEqual, testIPDefaultName)

			customIP := "255.255.255.255"
			os.Setenv(changeToEnvName(logLevelArg), "0") // 0 means debug level.
			os.Setenv(changeToEnvName(ipAddressArg), customIP)

			MustParseOnlyEnv()

			// Should be from environment.
			So(LogLevel(), ShouldEqual, logrus.DebugLevel)
			So(IPAddress(), ShouldEqual, customIP)
		})

		Convey("When some custom argument is defined", func() {
			// Register custom Argument.
			customArg := RegisterStringOption(customArgKey, customArgHelp, customArgDefault)

			Convey("Without parse it should be empty", func() {
				So(*customArg, ShouldEqual, "")
			})

			Convey("When we not defined any environment variable we should have default value after parse", func() {
				MustParseOnlyEnv()
				So(*customArg, ShouldEqual, customArgDefault)
			})

			Convey("When we define custom environment variable we should have custom value after parse", func() {
				customValue := "customContent"
				os.Setenv(changeToEnvName(customArgKey), customValue)

				MustParseOnlyEnv()
				So(*customArg, ShouldEqual, customValue)
			})
		})
	})
}
