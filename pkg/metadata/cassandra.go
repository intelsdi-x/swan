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
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/pkg/errors"
)

// CassandraConfig encodes the settings for connecting to the database.
type CassandraConfig struct {
	Address           string
	ConnectionTimeout time.Duration
	CreateKeyspace    bool
	IgnorePeerAddr    bool
	InitialHostLookup bool
	KeyspaceName      string
	Password          string
	Port              int
	SslCAPath         string
	SslCertPath       string
	SslEnabled        bool
	SslHostValidation bool
	SslKeyPath        string
	Timeout           time.Duration
	Username          string
}

// Cassandra is a helper struct which keeps the Cassandra session alive,
// holds the active configuration and the experiment id to tag the metadata with.
type Cassandra struct {
	experimentID string
	config       CassandraConfig
	session      *gocql.Session
}

// DefaultCassandraConfig applies the Cassandra settings from the command line flags and
// environment variables.
func DefaultCassandraConfig() CassandraConfig {
	return CassandraConfig{
		Address:           conf.CassandraAddress.Value(),
		ConnectionTimeout: time.Duration(conf.CassandraConnectionTimeout.Value()) * time.Second,
		CreateKeyspace:    conf.CassandraCreateKeyspace.Value(),
		IgnorePeerAddr:    conf.CassandraIgnorePeerAddr.Value(),
		InitialHostLookup: conf.CassandraInitialHostLookup.Value(),
		KeyspaceName:      conf.CassandraKeyspaceName.Value(),
		Password:          conf.CassandraPassword.Value(),
		Port:              conf.CassandraPort.Value(),
		SslCAPath:         conf.CassandraSslCAPath.Value(),
		SslCertPath:       conf.CassandraSslCertPath.Value(),
		SslEnabled:        conf.CassandraSslEnabled.Value(),
		SslHostValidation: conf.CassandraSslHostValidation.Value(),
		SslKeyPath:        conf.CassandraSslKeyPath.Value(),
		Timeout:           time.Duration(conf.CassandraTimeout.Value()) * time.Second,
		Username:          conf.CassandraUsername.Value(),
	}
}

// NewCassandra returns the Metadata helper from an experiment id and configuration.
func NewCassandra(experimentID string, config CassandraConfig) (Metadata, error) {
	metadata := &Cassandra{
		experimentID: experimentID,
		config:       config,
	}
	err := connect(metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func sslOptions(config CassandraConfig) *gocql.SslOptions {
	sslOptions := &gocql.SslOptions{
		EnableHostVerification: config.SslHostValidation,
	}

	if config.SslCAPath != "" {
		sslOptions.CaPath = config.SslCAPath
	}

	if config.SslCertPath != "" {
		sslOptions.CertPath = config.SslCertPath
	}

	if config.SslKeyPath != "" {
		sslOptions.KeyPath = config.SslKeyPath
	}

	return sslOptions
}

// getClusterConfig prepares configuration to Cassandra cluster.
func getClusterConfig(m *Cassandra) *gocql.ClusterConfig {
	cluster := gocql.NewCluster(m.config.Address)

	// TODO(niklas): make consistency configurable.
	cluster.Consistency = gocql.LocalOne
	cluster.SerialConsistency = gocql.LocalSerial

	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = m.config.ConnectionTimeout
	cluster.Timeout = m.config.Timeout
	cluster.IgnorePeerAddr = m.config.IgnorePeerAddr
	cluster.DisableInitialHostLookup = !m.config.InitialHostLookup

	return cluster
}

func createKeyspace(m *Cassandra, clusterConfig *gocql.ClusterConfig) error {
	session, err := clusterConfig.CreateSession()
	if err != nil {
		return errors.Wrap(err, "cannot create session for creating keyspace")
	}
	defer session.Close()

	query := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};", m.config.KeyspaceName)

	return errors.Wrap(session.Query(query).Exec(), "cannot create keyspace")

}

// connect creates a session to the Cassandra cluster. This function should only be called once.
func connect(m *Cassandra) error {
	cluster := getClusterConfig(m)
	cluster.Keyspace = m.config.KeyspaceName

	if m.config.Username != "" && m.config.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: m.config.Username,
			Password: m.config.Password,
		}
	}

	if m.config.SslEnabled {
		cluster.SslOpts = sslOptions(m.config)
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}

	m.session = session

	if m.config.CreateKeyspace {
		if err = createKeyspace(m, cluster); err != nil {
			return err
		}
	}

	if err = session.Query("CREATE TABLE IF NOT EXISTS metadata (experiment_id text, kind text, time timestamp, timeuuid TIMEUUID, metadata map<text,text>, PRIMARY KEY ((experiment_id), timeuuid),) WITH CLUSTERING ORDER BY (timeuuid DESC);").Exec(); err != nil {
		return err
	}

	return nil
}

// storeMap
func storeMap(m *Cassandra, metadata map[string]string, kind string) error {
	err := m.session.Query(`INSERT INTO metadata (experiment_id, kind, time, timeuuid, metadata) VALUES (?, ?, ?, ?, ?)`, m.experimentID, kind, time.Now(), gocql.TimeUUID(), metadata).Exec()
	return errors.Wrapf(err, "cannot publish metadata of kind %q", kind)
}

// Record stores a key and value and associates with the experiment id.
func (m *Cassandra) Record(key, value, kind string) error {
	metadata := map[string]string{}
	metadata[key] = value
	return storeMap(m, metadata, kind)
}

// RecordMap stores a key and value map and associates with the experiment id.
func (m *Cassandra) RecordMap(metadata map[string]string, kind string) error {
	return storeMap(m, metadata, kind)
}

// GetByKind retrive signle kind from the database.
// Returns error if no kind or too many groups found.
func (m *Cassandra) GetByKind(kind string) (map[string]string, error) {
	var metadata map[string]string

	maps := []map[string]string{}

	iter := m.session.Query(`SELECT metadata FROM metadata WHERE experiment_id = ? AND kind = ? ALLOW FILTERING`, m.experimentID, kind).Iter()
	for iter.Scan(&metadata) {
		maps = append(maps, metadata)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	// Make sure that only one map per experiment exists.
	if len(maps) != 1 {
		return nil, fmt.Errorf("Cannot retrieve metadata for experiment ID  %q and %q kind", m.experimentID, kind)
	}
	return maps[0], nil
}

// Clear deletes all metadata entries associated with the current experiment id.
func (m *Cassandra) Clear() error {
	if err := m.session.Query(`DELETE FROM metadata WHERE experiment_id = ?`,
		m.experimentID).Exec(); err != nil {
		return err
	}

	return nil
}
