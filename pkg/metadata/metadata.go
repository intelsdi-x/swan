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
	"fmt"

	"github.com/intelsdi-x/swan/pkg/conf"
)

// Predefined types of metadata.
// This selector allows to group metadata by their common characteristics.
// For instancje metadataKindFlags can be added to parameters passed to swan,
// metadataKindEnviron for environment variable and metadataKindPlatform for
// recorded platform characteristics like number of CPUs and so on.
// Note that metadataKind is just a string and each experiment can define
// it own personalized types of metadata.
const (
	TypeEmpty    = ""
	TypeFlags    = "flags"
	TypeEnviron  = "environ"
	TypePlatform = "platform"
)

// Metadata interface defines methods which must be supported by DB backend
type Metadata interface {
	// Record stores a key and value and associates with the experiment id.
	Record(key string, value string, kind string) error
	// RecordMap stores a key and value map and associates with the experiment id.
	RecordMap(metadata map[string]string, kind string) error
	// GetByKind retrives single metadata type from the database.
	// Returns error if no kind or too many groups found.
	GetByKind(kind string) (map[string]string, error)
	// Clear deletes all metadata entries associated with the current experiment id.
	Clear() error
}

// NewDefault initialize metadata object which is configured via env. variable.
func NewDefault(experimentID string) (Metadata, error) {
	if conf.DefaultMetadataDB.Value() == "cassandra" {
		return NewCassandra(experimentID, DefaultCassandraConfig())
	}

	if conf.DefaultMetadataDB.Value() == "influxdb" {
		return NewInfluxDB(experimentID, DefaultInfluxDBConfig())
	}

	return nil, fmt.Errorf("Unsupported database for metadata: %s", conf.DefaultMetadataDB.Value())
}
