// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conf

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/pkg/errors"
)

var (
	// Default flags and values.
	logLevelFlag = NewStringFlag(
		"log_level",
		"Log level for Swan: debug, info, warn, error, fatal, panic",
		"info",
	)
)

const envPrefix = "SWAN_"

// LogLevel returns configured logLevel from input option or env variable.
// If it cannot parse the log level, it returns default value.
func LogLevel() (logrus.Level, error) {
	level, err := logrus.ParseLevel(logLevelFlag.Value())
	if err != nil {
		return logrus.PanicLevel, errors.Wrap(err, "cannot parse 'log' level flag")
	}
	return level, nil
}

// LoadConfig from given file the is simple environment format.
// Description:
// - '#' indicates means comment,
// - every other line containing '=' is splited as key and value for environment.
func LoadConfig(filename string) error {
	config, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "cannot load config file: %s: %v", filename, err)
	}
	for _, line := range strings.Split(string(config), "\n") {

		if !strings.HasPrefix(line, "#") && strings.Contains(line, "=") {
			fields := strings.Split(line, "=")
			os.Setenv(envPrefix+fields[0], fields[1])
		}
	}
	return nil
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
func DumpConfigMap(flagMap map[string]string) string {
	buffer := &bytes.Buffer{}

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

		fmt.Fprintf(buffer, "%s=%v\n", strings.ToUpper(fd.Name), value)
	}

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
