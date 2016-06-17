// conf is a helper for SWAN configuration for both command line interface
// and environment variables.
// It gives ability to register arguments which will be fetched from
// CLI input OR environment variable.
// By default it registers following options:
// <SWAN_LOG> -l --log <Log level for Swan 0:debug; 1:info; 2:warn; 3:error; 4:fatal, 5:panic> Default: 3
//
// When `ParseEnv` is executed, only the environment arguments are parsed and
// ready to be used in `promises` variables.
// `ParseOnlyEnv` can be run multiple times.
//
// When `ParseFlagAndEnv` is executed, the arguments from both CLI and Env are parsed.
// In case of --help option - it prints help.
// It's recommend to run it only once, since to have `conf` with registered all needed options from
// system. When help option is executed it will then show whole overview of the Swan configuration.

package conf

import (
	"github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
)

var (
	app = kingpin.New("swan", "No help available")
	// Default flags and values.
	logLevelFlag = NewStringFlag(
		"log",
		"Log level for Swan: debug, info, warn, error, fatal, panic",
		"error", // Default Error log level.
	)
	isEnvParsed = false
)

// SetHelpPath sets the help message for CLI rendering the file from given file.
// We need to expose this function so other packages can set the app help.
func SetHelpPath(readmePath string) {
	readmeData, err := ioutil.ReadFile(readmePath)
	if err != nil {
		panic(err)
	}
	app.Help = string(readmeData)[:]
}

// SetAppName sets application name for CLI output.
// We need to expose this function so other packages can set the app name.
func SetAppName(name string) {
	app.Name = name
}

// LogLevel returns configured logLevel from input option or env variable.
// If it cannot parse the log level, it returns default value.
func LogLevel() logrus.Level {
	level, err := logrus.ParseLevel(logLevelFlag.Value())
	if err == nil {
		return level
	}

	level, err = logrus.ParseLevel(logLevelFlag.defaultValue)
	if err == nil {
		return level

	}

	// Programmer error.
	panic(err)
}

// AppName returns specified app name.
func AppName() string {
	return app.Name
}

// ParseFlags parse both the command line flags of the process and
// environment variables.
func ParseFlags() error {
	_, err := app.Parse(os.Args[1:])
	if err == nil {
		isEnvParsed = true
	}
	return err
}

// ParseEnv parse the environment for arguments.
func ParseEnv() error {
	_, err := app.Parse([]string{})
	if err == nil {
		isEnvParsed = true
	}
	return err
}
