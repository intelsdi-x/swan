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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
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
		panic(errors.Wrapf(err, "reading %s failed", readmePath))
	}
	app.Help = string(readmeData)[:]
}

// SetHelp sets the help message for the CLI.
// We need to expose this function so other packages can set the app help.
func SetHelp(help string) {
	app.Help = help
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
	panic(errors.Wrap(err, "parsing log level failed"))
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
		return nil
	}

	return errors.Wrapf(err, "could not parse command line flags")
}

// ParseEnv parse the environment for arguments.
func ParseEnv() error {
	_, err := app.Parse([]string{})
	if err == nil {
		isEnvParsed = true
		return nil
	}

	return errors.Wrapf(err, "could not parse enviroment flags")
}

// getFlagsDefinition returns current, default, keys and descrition for every flag.
// Notes: order is important because it logically groups flags.
func getFlagsDefinition() (flags []struct{ Name, Value, Default, Help string }) {

	for _, flag := range app.Model().Flags {

		// Skip kingpin builtin flags that aren't compatibile with environment based configuration.
		if strings.Contains(flag.Name, "-") {
			continue
		}

		// Returned values are basic types (string, int) or time.Duration and then serialized to string.
		var value interface{} // golang native type

		// First handle pkg/conf swan internal flags implmentation.
		if slv, ok := flag.Value.(*StringListValue); ok {
			value = slv.String()
		} else {
			// Use reflection to extract value hidden in non exported kingpin implmentation.

			// Extract reflect.Value from kingpin interface (kingpin.Value).
			reflectValue := reflect.ValueOf(flag.Value)

			// Dereference point from reflect.Value.
			// Value represent a pointer to something lke kingping.boolValue or kingping.stringValue, so extrac the _Value struct itself.
			elem := reflectValue.Elem()

			// Basing on underlaying type convert to native type.
			// Laws of reflection:
			// "The second property is that the Kind of a reflection object describes the underlying type, not the static type."
			switch elem.Kind() {

			case reflect.Int64, reflect.Int:
				// Special case for duration flag that is not stored in
				value = time.Duration(elem.Int())

			case reflect.Struct:

				// Get field that is used in kingpin to store value (pointer)
				// and dereference pointer.
				field := elem.FieldByName("v")
				valueInField := field.Elem()

				// Check the underlying type of value stored in v.
				switch valueInField.Kind() {

				case reflect.String:
					value = valueInField.String()

				case reflect.Bool:
					value = valueInField.Bool()

				case reflect.Int64, reflect.Int:
					value = valueInField.Int()

				default:
					fmt.Sprintf("unhandled flag %s kind=%s", flag.Name, elem.Kind())
				}
			}
		}

		flags = append(flags, struct{ Name, Value, Default, Help string }{
			Name:    flag.Name,
			Help:    flag.Help,
			Default: strings.Join(flag.Default, ","),
			Value:   fmt.Sprintf("%v", value), // serialize value to String.
		})
	}

	return flags
}

// DumpConfig dumps evironment based configuration with current values of flags.
func DumpConfig() string {
	return DumpConfigMap(nil)
}

// DumpConfigMap dumps evironment based configuration with current values overwritten by given flagMap.
// Includes "allexport" directives for bash.
func DumpConfigMap(flagMap map[string]string) string {
	buffer := &bytes.Buffer{}

	buffer.WriteString("# Export are values.\n")
	buffer.WriteString("set -o allexport\n")

	for _, fd := range getFlagsDefinition() {

		fmt.Fprintf(buffer, "\n# %s\n", fd.Help)
		if fd.Default != "" {
			fmt.Fprintf(buffer, "# Default: %s\n", fd.Default)
		}

		// Override current values with provided from flagMap.
		value := fd.Value
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
		flagsMap[flag.Name] = flag.Value
	}
	return flagsMap
}
