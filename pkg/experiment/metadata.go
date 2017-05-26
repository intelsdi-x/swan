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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/pkg/errors"
)

const (
	metadataKindEmpty    = ""
	metadataKindFlags    = "flags"
	metadataKindEnviron  = "environ"
	metadataKindPlatform = "platform"
)

// MetadataConfig encodes the settings for connecting to the database.
type MetadataConfig struct {
	CassandraAddress           string
	CassandraConnectionTimeout time.Duration
	CassandraCreateKeyspace    bool
	CassandraIgnorePeerAddr    bool
	CassandraInitialHostLookup bool
	CassandraKeyspaceName      string
	CassandraPassword          string
	CassandraPort              int
	CassandraSslCAPath         string
	CassandraSslCertPath       string
	CassandraSslEnabled        bool
	CassandraSslHostValidation bool
	CassandraSslKeyPath        string
	CassandraTimeout           time.Duration
	CassandraUsername          string
}

// DefaultMetadataConfig returns a setup which use a Cassandra cluster running on localhost
// without any authentication or encryption.
func DefaultMetadataConfig() MetadataConfig {
	return MetadataConfig{
		CassandraAddress:           "127.0.0.1",
		CassandraConnectionTimeout: 0,
		CassandraCreateKeyspace:    true,
		CassandraIgnorePeerAddr:    false,
		CassandraInitialHostLookup: true,
		CassandraKeyspaceName:      "swan",
		CassandraPassword:          "",
		CassandraPort:              9042,
		CassandraSslCAPath:         "",
		CassandraSslCertPath:       "",
		CassandraSslEnabled:        false,
		CassandraSslHostValidation: false,
		CassandraSslKeyPath:        "",
		CassandraTimeout:           0,
		CassandraUsername:          "",
	}
}

// MetadataConfigFromFlags applies the Cassandra settings from the command line flags and
// environment variables.
func MetadataConfigFromFlags() MetadataConfig {
	return MetadataConfig{
		CassandraAddress:           conf.CassandraAddress.Value(),
		CassandraConnectionTimeout: time.Duration(conf.CassandraConnectionTimeout.Value()) * time.Second,
		CassandraCreateKeyspace:    conf.CassandraCreateKeyspace.Value(),
		CassandraIgnorePeerAddr:    conf.CassandraIgnorePeerAddr.Value(),
		CassandraInitialHostLookup: conf.CassandraInitialHostLookup.Value(),
		CassandraKeyspaceName:      conf.CassandraKeyspaceName.Value(),
		CassandraPassword:          conf.CassandraPassword.Value(),
		CassandraPort:              conf.CassandraPort.Value(),
		CassandraSslCAPath:         conf.CassandraSslCAPath.Value(),
		CassandraSslCertPath:       conf.CassandraSslCertPath.Value(),
		CassandraSslEnabled:        conf.CassandraSslEnabled.Value(),
		CassandraSslHostValidation: conf.CassandraSslHostValidation.Value(),
		CassandraSslKeyPath:        conf.CassandraSslKeyPath.Value(),
		CassandraTimeout:           time.Duration(conf.CassandraTimeout.Value()) * time.Second,
		CassandraUsername:          conf.CassandraUsername.Value(),
	}
}

// MetadataMap encodes the key value pairs to be stored in Cassandra.
type MetadataMap map[string]string

// Metadata is a helper struct which keeps the Cassandra session alive, holds the active configuration
// and the experiment id to tag the metadata with.
type Metadata struct {
	experimentID string
	config       MetadataConfig
	session      *gocql.Session
}

// NewMetadata returns the Metadata helper from an experiment id and configuration.
func NewMetadata(experimentID string, config MetadataConfig) (*Metadata, error) {
	metadata := &Metadata{
		experimentID: experimentID,
		config:       config,
	}
	err := metadata.Connect()
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func sslOptions(config MetadataConfig) *gocql.SslOptions {
	sslOptions := &gocql.SslOptions{
		EnableHostVerification: config.CassandraSslHostValidation,
	}

	if config.CassandraSslCAPath != "" {
		sslOptions.CaPath = config.CassandraSslCAPath
	}

	if config.CassandraSslCertPath != "" {
		sslOptions.CertPath = config.CassandraSslCertPath
	}

	if config.CassandraSslKeyPath != "" {
		sslOptions.KeyPath = config.CassandraSslKeyPath
	}

	return sslOptions
}

// connect creates a session to the Cassandra cluster. This function should only be called once.
func (m *Metadata) connect() *gocql.ClusterConfig {
	cluster := gocql.NewCluster(m.config.CassandraAddress)

	// TODO(niklas): make consistency configurable.
	cluster.Consistency = gocql.LocalOne
	cluster.SerialConsistency = gocql.LocalSerial

	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = m.config.CassandraConnectionTimeout
	cluster.Timeout = m.config.CassandraTimeout
	cluster.IgnorePeerAddr = m.config.CassandraIgnorePeerAddr
	cluster.DisableInitialHostLookup = !m.config.CassandraInitialHostLookup

	return cluster
}

func (m *Metadata) createKeyspace() error {
	connectionConfiguration := m.connect()
	session, err := connectionConfiguration.CreateSession()
	defer session.Close()
	if err != nil {
		return errors.Wrap(err, "cannot create session for creating keyspace")
	}

	query := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};", m.config.CassandraKeyspaceName)

	return errors.Wrap(session.Query(query).Exec(), "cannot create keyspace")

}

// Connect creates a session to the Cassandra cluster. This function should only be called once.
func (m *Metadata) Connect() error {
	cluster := m.connect()
	cluster.Keyspace = m.config.CassandraKeyspaceName

	if m.config.CassandraUsername != "" && m.config.CassandraPassword != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: m.config.CassandraUsername,
			Password: m.config.CassandraPassword,
		}
	}

	if m.config.CassandraSslEnabled {
		cluster.SslOpts = sslOptions(m.config)
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}

	m.session = session

	if m.config.CassandraCreateKeyspace {
		if err = m.createKeyspace(); err != nil {
			return err
		}
	}

	if err = session.Query("CREATE TABLE IF NOT EXISTS metadata (experiment_id text, kind text, time timestamp, timeuuid TIMEUUID, metadata map<text,text>, PRIMARY KEY ((experiment_id), timeuuid),) WITH CLUSTERING ORDER BY (timeuuid DESC);").Exec(); err != nil {
		return err
	}

	return nil
}

// storeMap
func (m *Metadata) storeMap(metadata MetadataMap, kind string) error {
	err := m.session.Query(`INSERT INTO metadata (experiment_id, kind, time, timeuuid, metadata) VALUES (?, ?, ?, ?, ?)`, m.experimentID, kind, time.Now(), gocql.TimeUUID(), metadata).Exec()
	return errors.Wrapf(err, "cannot publish metadata of kind %q", kind)
}

// Record stores a key and value and associates with the experiment id.
func (m *Metadata) Record(key string, value string) error {
	metadata := MetadataMap{}
	metadata[key] = value
	return m.storeMap(metadata, metadataKindEmpty)
}

// RecordMap stores a key and value map and associates with the experiment id.
func (m *Metadata) RecordMap(metadata MetadataMap) error {
	return m.storeMap(metadata, metadataKindEmpty)
}

//RecordRuntimeEnv store experiment environment information in Cassandra.
func (m *Metadata) RecordRuntimeEnv(experimentStart time.Time) error {
	// Store configuration.
	err := m.recordFlags()
	if err != nil {
		return err
	}

	// Store SWAN_ environment configuration.
	err = m.recordEnv(conf.EnvironmentPrefix)
	if err != nil {
		return err
	}

	// Store host and time in metadata.
	hostname, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "cannot retrieve hostname")
	}
	// Store hostname and start time.
	err = m.RecordMap(map[string]string{"time": experimentStart.Format(time.RFC822Z), "host": hostname})
	if err != nil {
		return err
	}

	// Store hardware & OS details.
	err = m.recordPlatformMetrics()
	if err != nil {
		return err
	}

	return nil
}

// recordFlags saves whole flags based configuration in the metadata information.
func (m *Metadata) recordFlags() error {
	metadata := conf.GetFlags()
	return m.storeMap(metadata, metadataKindFlags)
}

// recordEnv adds all OS Environment variables that starts with prefix 'prefix'
// in the metadata information
func (m *Metadata) recordEnv(prefix string) error {
	metadata := MetadataMap{}
	// Store environment configuration.
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			fields := strings.SplitN(env, "=", 2)
			metadata[fields[0]] = fields[1]
		}
	}
	return m.storeMap(metadata, metadataKindEnviron)
}

// recordPlatformMetrics stores platform specific metadata.
// Platform metrics are metadataKindPlatform type.
func (m *Metadata) recordPlatformMetrics() error {
	platformMetrics := GetPlatformMetrics()
	return m.storeMap(platformMetrics, metadataKindPlatform)
}

// Get retrieves all metadata maps from the database.
func (m *Metadata) Get() ([]MetadataMap, error) {
	var metadata MetadataMap

	out := []MetadataMap{}

	iter := m.session.Query(`SELECT metadata FROM metadata WHERE experiment_id = ?`, m.experimentID).Iter()
	for iter.Scan(&metadata) {
		out = append(out, metadata)
	}
	if err := iter.Close(); err != nil {
		return []MetadataMap{}, err
	}

	return out, nil
}

// GetGroup retrive signle kind from the database.
// Returns error if no kind or too many groups found.
func (m *Metadata) GetGroup(kind string) (MetadataMap, error) {
	var metadata MetadataMap

	maps := []MetadataMap{}

	iter := m.session.Query(`SELECT metadata FROM metadata WHERE experiment_id = ? AND kind = ? ALLOW FILTERING`, m.experimentID, kind).Iter()
	for iter.Scan(&metadata) {
		maps = append(maps, metadata)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	// Make sure that only one map withing experiment exists.
	if len(maps) != 1 {
		return nil, fmt.Errorf("Cannot retrieve metadata for experiment ID  %q and %q kind", m.experimentID, kind)
	}
	return maps[0], nil
}

// Clear deletes all metadata entries associated with the current experiment id.
func (m *Metadata) Clear() error {
	if err := m.session.Query(`DELETE FROM metadata WHERE experiment_id = ?`,
		m.experimentID).Exec(); err != nil {
		return err
	}

	return nil
}
