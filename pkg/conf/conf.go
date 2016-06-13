package conf

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"strings"
)

// conf is a helper for SWAN configuration for both command line interface
// and environment variables.
// It gives ability to register arguments which will be fetched from
// CLI input OR environment variable.
// By default it registers following options:
// <SWAN_LOG> -l --log <Log level for Swan 0:debug; 1:info; 2:warn; 3:error; 4:fatal, 5:panic> Default: 3
// <SWAN_IP> -i --ip <IP of interface for Swan workloads services to listen on> Default: 127.0.0.1
//
// When `MustParseOnlyEnv` is executed, only the environment arguments are parsed and
// ready to be used in `promises` variables.
// `MustParseOnlyEnv` can be run multiple times.
//
// When `MustParseCliAndEnv` is executed, the arguments from both CLI and Env are parsed.
// In case of --help option - it prints help.
// It's recommend to run it only once, since to have `conf` with registered all needed options from
// system. When help option is executed it will then show whole overview of the Swan configuration.

var (
	// Default flags and values.
	logLevelArg      = "log" // Short 'l'.
	ipAddressArg     = "ip"  // Short 'i'.
	defaultLogLevel  = "3"   // Error log level.
	defaultIPAddress = "127.0.0.1"

	// Package vars.
	app       *kingpin.Application
	logLevel  *int
	iPAddress *string
)

// changeToEnvName creates environment variable name by making it uppercase and adding SWAN prefix.
// For instance: "cassandra_host" will be "SWAN_CASSANDRA_HOST".
func changeToEnvName(name string) string {
	return fmt.Sprintf("%s_%s", "SWAN", strings.ToUpper(name))
}

// init constructs config with default help message and application name.
// It also defines two important, default options like logLevel and IP of remote interface.
// TODO(bp): Decide if we want specifying IP vs hostnames and deploy proper hosts into /etc/hosts.
func init() {
	app = kingpin.New("swan", "No help available")
	logLevel = app.Flag(
		logLevelArg, "Log level for Swan 0:debug; 1:info; 2:warn; 3:error; 4:fatal, 5:panic",
	).OverrideDefaultFromEnvar(changeToEnvName(logLevelArg)).Default(defaultLogLevel).Short('l').Int()

	iPAddress = app.Flag(
		ipAddressArg, "IP of interface for Swan workloads services to listen on",
	).OverrideDefaultFromEnvar(changeToEnvName(ipAddressArg)).Default(defaultIPAddress).Short('i').String()

	MustParseOnlyEnv()
}

// SetHelpPath sets the help message for CLI rendering the file from given file.
func SetHelpPath(readmePath string) {
	readmeData, err := ioutil.ReadFile(readmePath)
	if err != nil {
		panic(err)
	}
	app.Help = string(readmeData)[:]
}

// SetAppName sets application name for CLI output.
func SetAppName(name string) {
	app.Name = name
}

// LogLevel returns configured logLevel from input option or env variable.
func LogLevel() logrus.Level {
	// Since the logrus defines levels as iota enum
	// (https://github.com/Sirupsen/logrus/blob/master/logrus.go#L57)
	// We just need to roll over the enum to achieve our API (0:debug, 5:Panic)
	return logrus.AllLevels[len(logrus.AllLevels)-(*logLevel+1)]
}

// IPAddress returns IP which will be specified for workload services as endpoints.
// TODO(bp): In future we can specify here the experiment host Address. This will be available only
// when we will be ready to make remote isolations.
func IPAddress() string {
	return *iPAddress
}

// AppName returns specified app name.
func AppName() string {
	return app.Name
}

// RegisterStringOption registers given option in form of name, help msg and default value
// as optional string argument in CLI. It defines also overrideDefaultFromEnvar rule for this
// argument.
func RegisterStringOption(name string, help string, defaultVal string) *string {
	return app.Flag(
		name, help).Default(defaultVal).OverrideDefaultFromEnvar(changeToEnvName(name)).String()
}

// MustParseFlagAndEnv parse both the command line flags of the process and
// environment variables.
// It panics in case of error.
func MustParseFlagAndEnv() {
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}
}

// MustParseOnlyEnv parse the environment for arguments.
// It panics in case of error.
func MustParseOnlyEnv() {
	_, err := app.Parse([]string{})
	if err != nil {
		panic(err)
	}
}
