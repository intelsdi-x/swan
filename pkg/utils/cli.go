package utils

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"strings"
)

var (
	CassandraHostArg     = "cassandra_host"
	LoadGeneratorHostArg = "load_generator_host"
	SnapHostArg          = "snap_host"
	logLevelArg          = "log" // short 'l'
)

type Cli struct {
	AppName    string
	app        *kingpin.Application
	stringArgs map[string]*string
	logLevel   *int
}

// Environment variable from "cassandra_host" will be "SWAN_CASSANDRA_HOST"
func changeToEnvName(name string) string {
	return fmt.Sprintf("%s_%s", "SWAN", strings.ToUpper(name))
}

func NewCliWithReadme(appName string, readme string) *Cli {
	app := kingpin.New(appName, readme)
	logLevel := app.Flag(
		logLevelArg, "Log level for Swan 0:debug; 1:info; 2:warn; 3:error; 4:fatal, 5:panic",
	).OverrideDefaultFromEnvar(changeToEnvName(logLevelArg)).Default("3").Short('l').Int()

	return &Cli{
		AppName:    appName,
		app:        app,
		stringArgs: make(map[string]*string),
		logLevel:   logLevel,
	}
}

func (c *Cli) addRequiredArgWithEnv(name string, help string, defaultVal string) {
	c.stringArgs[name] = c.app.Flag(
		name, help).Default(defaultVal).OverrideDefaultFromEnvar(changeToEnvName(name)).String()
}

func (c *Cli) Get(argName string) string {
	arg, ok := c.stringArgs[argName]
	if !ok {
		panic(fmt.Sprintf("%s was not spefied as input argument for this CLI", argName))
	}

	return *arg
}

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

func (c *Cli) AddCassandraHostArg() *Cli {
	c.addRequiredArgWithEnv(CassandraHostArg, "Host for Cassandra DB", "127.0.0.1")
	return c
}

func (c *Cli) AddLoadGeneratorHostArg() *Cli {
	c.addRequiredArgWithEnv(LoadGeneratorHostArg, "Host for Load Generator", "127.0.0.1")
	return c
}

func (c *Cli) AddSnapHostArg() *Cli {
	c.addRequiredArgWithEnv(SnapHostArg, "Host for Snap", "127.0.0.1")
	return c
}

func (c *Cli) MustParse() *Cli {
	_, err := c.app.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	return c
}
