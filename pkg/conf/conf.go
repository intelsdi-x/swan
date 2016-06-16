// conf is a helper for SWAN configuration for both command line interface
// and environment variables.
// It gives ability to register arguments which will be fetched from
// CLI input OR environment variable.
// By default it registers following options:
// <SWAN_LOG> -l --log <Log level for Swan 0:debug; 1:info; 2:warn; 3:error; 4:fatal, 5:panic> Default: 3
// <SWAN_IP> -i --ip <IP of interface for Swan workloads services to listen on> Default: 127.0.0.1
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
	logLevelFlag = NewRegisteredIntFlag(
		"log",
		"Log level for Swan 0:debug; 1:info; 2:warn; 3:error; 4:fatal, 5:panic",
		"3", // Default Error log level.
	)
	ipAddressFlag = NewRegisteredStringFlag(
		"ip",
		"IP of interface for Swan workloads services to listen on",
		"127.0.0.1",
	)
)

// init parses.
func init() {
	err := ParseEnv()
	if err != nil {
		panic(err)
	}
}

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
func LogLevel() logrus.Level {
	// Since the logrus defines levels as iota enum
	// (https://github.com/Sirupsen/logrus/blob/master/logrus.go#L57)
	// We just need to roll over the enum to achieve our API (0:debug, 5:Panic)
	return logrus.AllLevels[len(logrus.AllLevels)-(logLevelFlag.Value()+1)]
}

// IPAddress returns IP which will be specified for workload services as endpoints.
// TODO(bp): In future we can specify here the experiment host Address. This will be available only
// when we will be ready to make remote isolations.
func IPAddress() string {
	return ipAddressFlag.Value()
}

// AppName returns specified app name.
func AppName() string {
	return app.Name
}

// ParseFlagAndEnv parse both the command line flags of the process and
// environment variables.
func ParseFlagAndEnv() error {
	_, err := app.Parse(os.Args[1:])
	return err
}

// ParseEnv parse the environment for arguments.
// NOTE: Make sure you run it after flag registration.
func ParseEnv() error {
	_, err := app.Parse([]string{})
	return err
}
