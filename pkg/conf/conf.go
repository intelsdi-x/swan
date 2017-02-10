package conf

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

var (
	// Default flags and values.
	logLevelFlag = NewStringFlag(
		"log",
		"Log level for Swan: debug, info, warn, error, fatal, panic",
		"error", // Default Error log level.
	)
	isEnvParsed = false
)

// LogLevel returns configured logLevel from input option or env variable.
// If it cannot parse the log level, it returns default value.
func LogLevel() logrus.Level {
	level, err := logrus.ParseLevel(logLevelFlag.Value())
	if err == nil {
		return level
	}
	// Programmer error.
	panic(errors.Wrap(err, "parsing log level failed"))
}

// ParseFlags parse both the command line flags of the process and
// environment variables. Command line flags override values from
// environment.
func ParseFlags() {
	flag.VisitAll(func(flag *flag.Flag) {
		value := os.Getenv(envName(flag.Name))
		if value != "" {
			flag.Value.Set(value)
		}
	})
	flag.Parse()
}

// getFlagsDefinition returns definition of all flags for internal purpose.
func getFlagsDefinition() (flags []*flag.Flag) {
	// Get all flags.
	flagsMap := map[string]*flag.Flag{}
	flag.VisitAll(func(flag *flag.Flag) {
		flagsMap[flag.Name] = flag
	})

	// Filter by registered names.
	for _, name := range flagNames {
		if flag, ok := flagsMap[name]; ok {
			flags = append(flags, flag)
		}
	}
	return
}

// DumpConfig dumps environment based configuration with current values of flags.
func DumpConfig() string {
	return DumpConfigMap(nil)
}

// DumpConfigMap dumps environment based configuration with current values overwritten by given flagMap.
// Includes "allexport" directives for bash.
func DumpConfigMap(flagMap map[string]string) string {
	buffer := &bytes.Buffer{}

	buffer.WriteString("# Export are values.\n")
	buffer.WriteString("set -o allexport\n")

	for _, fd := range getFlagsDefinition() {

		fmt.Fprintf(buffer, "\n# %s\n", fd.Usage)
		if fd.DefValue != "" {
			fmt.Fprintf(buffer, "# Default: %s\n", fd.DefValue)
		}

		// Override current values with provided from flagMap.
		value := fd.Value.String()
		if mapValue, ok := flagMap[fd.Name]; ok {
			value = mapValue
		}

		fmt.Fprintf(buffer, "SWAN_%s=%v\n", strings.ToUpper(fd.Name), value)
	}

	buffer.WriteString("set +o allexport")
	return buffer.String()
}

// GetFlags returns flags as map with current values.
func GetFlags() map[string]string {
	flagsMap := map[string]string{}
	for _, flag := range getFlagsDefinition() {
		flagsMap[flag.Name] = flag.Value.String()
	}
	return flagsMap
}
