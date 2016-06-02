package utils

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"strings"
)

var (
	logLevelArg  = "log" // short 'l'
	ipAddressArg = "ip"  // short 'i'
)

// Cli is a helper for SWAN command line interface.
// It gives ability to register arguments which will be fetched from
// CLI input of environment variable.
// By default it gives following options:
// -l --log <Log level for Swan 0:debug; 1:info; 2:warn; 3:error; 4:fatal, 5:panic> Default: 3
// -i --ip <IP of interface for Swan workloads services to listen on> Default: 127.0.0.1
// --help prints help.
type Cli struct {
	AppName   string
	app       *kingpin.Application
	logLevel  *int
	iPAddress *string
}

// Environment variable from "cassandra_host" will be "SWAN_CASSANDRA_HOST".
func changeToEnvName(name string) string {
	return fmt.Sprintf("%s_%s", "SWAN", strings.ToUpper(name))
}

// NewCliWithReadme constructs CLI where help will print README file in raw format.
// It also defines two important, default options like logLevel and IP of remote interface.
// TODO(bp): Decide if we want specifying IP vs hostnames and deploy proper hosts into /etc/hosts.
func NewCliWithReadme(appName string, readmePath string) *Cli {
	readmeData, err := ioutil.ReadFile(readmePath)
	if err != nil {
		panic(err)
	}

	app := kingpin.New(appName, string(readmeData)[:])
	logLevel := app.Flag(
		logLevelArg, "Log level for Swan 0:debug; 1:info; 2:warn; 3:error; 4:fatal, 5:panic",
	).OverrideDefaultFromEnvar(changeToEnvName(logLevelArg)).Default("3").Short('l').Int()

	// TODO(bp): Decide if we want specifying IP vs
	// hostnames and deploy proper hosts into /etc/hosts.
	iPAddress := app.Flag(
		ipAddressArg, "IP of interface for Swan workloads services to listen on",
	).OverrideDefaultFromEnvar(changeToEnvName(ipAddressArg)).Default("127.0.0.1").Short('i').String()

	return &Cli{
		AppName:   appName,
		app:       app,
		logLevel:  logLevel,
		iPAddress: iPAddress,
	}
}

// RegisterStringArgWithEnv registers given option in form of name, help msg and default value
// as optional string argument in CLI. It defines also overrideDefaultFromEnvar rule for this
// argument.
func (c *Cli) RegisterStringArgWithEnv(name string, help string, defaultVal string) *string {
	return c.app.Flag(
		name, help).Default(defaultVal).OverrideDefaultFromEnvar(changeToEnvName(name)).String()
}

// LogLevel returns configured logLevel from input arg or env variable.
func (c *Cli) LogLevel() logrus.Level {
	// Since the logrus defines levels as iota enum in such form:
	// PanicLevel Level = iota
	// FatalLevel
	// ErrorLevel
	// WarnLevel
	// InfoLevel
	// DebugLevel
	// We just need to roll over the enum to achieve our API (0:debug, 5:Panic)
	return logrus.AllLevels[len(logrus.AllLevels)-(*c.logLevel+1)]
}

// IPAddress returns IP which will be specified for workload services as endpoints.
// TODO(bp): In future we can specify here the experiment host Address. This will be available only
// when we will be ready to make remote isolations.
func (c *Cli) IPAddress() string {
	return *c.iPAddress
}

// MustParse parse the command line argument of the process.
// It panics in case of error.
func (c *Cli) MustParse() *Cli {
	_, err := c.app.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	return c
}
