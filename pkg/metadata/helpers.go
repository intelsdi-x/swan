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

package metadata

import (
	"os"
	"strings"
	"time"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/pkg/errors"
)

//RecordRuntimeEnv store experiment environment information in Cassandra.
func RecordRuntimeEnv(metadata Metadata, experimentStart time.Time) error {
	// Store configuration.
	err := recordFlags(metadata)
	if err != nil {
		return err
	}

	// Store SWAN_ environment configuration.
	err = recordEnv(metadata, conf.EnvironmentPrefix)
	if err != nil {
		return err
	}

	// Store host and time in metadata.
	hostname, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "cannot retrieve hostname")
	}
	// Store hostname and start time.
	err = metadata.RecordMap(map[string]string{"time": experimentStart.Format(time.RFC822Z), "host": hostname}, TypeEmpty)
	if err != nil {
		return err
	}

	// Store hardware & OS details.
	err = recordPlatformMetrics(metadata)
	if err != nil {
		return err
	}

	return nil
}

// recordFlags saves whole flags based configuration in the metadata information.
func recordFlags(metadata Metadata) error {
	flags := conf.GetFlags()
	return metadata.RecordMap(flags, TypeFlags)
}

// recordEnv adds all OS Environment variables that starts with prefix 'prefix'
// in the metadata information
func recordEnv(metadata Metadata, prefix string) error {
	envMetadata := map[string]string{}
	// Store environment configuration.
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			fields := strings.SplitN(env, "=", 2)
			envMetadata[fields[0]] = fields[1]
		}
	}
	return metadata.RecordMap(envMetadata, TypeEnviron)
}

// recordPlatformMetrics stores platform specific metadata.
// Platform metrics are metadataKindPlatform type.
func recordPlatformMetrics(metadata Metadata) error {
	platformMetrics := GetPlatformMetrics()
	return metadata.RecordMap(platformMetrics, TypePlatform)
}
