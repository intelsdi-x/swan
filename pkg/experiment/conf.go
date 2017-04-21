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

package experiment

import (
	"flag"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
)

var (
	// Flags are defined using directly go native "flag" package to not be registered as experiment configuration.
	loadConfig             = flag.String("config", "", "Load configuration from file")
	dumpConfig             = flag.Bool("config-dump", false, "Dump configuration as environment script.")
	dumpConfigExperimentID = flag.String("config-dump-experiment-id", "", "Dump configuration based on experiment ID.")
)

// Configure handles configuration parsing, generation and restoration based on config-* flags.
// Note: exits if configuration generation was requested.
// This function must reside in experiment package because depends on metadata access.
// Returns information about current log level.
func Configure() bool {

	// Load config from file if requested.
	flag.Parse()
	if *loadConfig != "" {
		err := conf.LoadConfig(*loadConfig)
		errutil.Check(err)
	}

	// Parse extended flags again using environment.
	err := conf.ParseFlags()
	errutil.Check(err)

	// Setup log level accordingly.
	level, err := conf.LogLevel()
	errutil.Check(err)
	logrus.SetLevel(level)

	if *dumpConfig {
		previousExperimentID := *dumpConfigExperimentID
		if previousExperimentID != "" {
			metadata, err := NewMetadata(previousExperimentID, MetadataConfigFromFlags())
			errutil.Check(err)
			flags, err := metadata.GetGroup("flags")
			errutil.Check(err)
			fmt.Println(conf.DumpConfigMap(flags))
		} else {
			fmt.Println(conf.DumpConfig())
		}
		os.Exit(0)
	}
	return level == logrus.ErrorLevel
}
