package conf

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
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
func LogLevel() (logrus.Level, error) {
	level, err := logrus.ParseLevel(logLevelFlag.Value())
	if err != nil {
		return logrus.PanicLevel, errors.Wrap(err, "cannot parse 'log' level flag")
	}
	return level, nil
}

// ParseFlags parse both the command line flags of the process and
// environment variables. Command line flags override values from
// environment.
func ParseFlags() error {
	var errCollection errcollection.ErrorCollection
	flag.VisitAll(func(flag *flag.Flag) {
		value := os.Getenv(envName(flag.Name))
		if value != "" {
			err := flag.Value.Set(value)
			if err != nil {
				errCollection.Add(errors.Wrapf(err, "cannot parse %q flag", flag.Name))
			}
		}
	})
	flag.Parse()
	return errCollection.GetErrIfAny()
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
