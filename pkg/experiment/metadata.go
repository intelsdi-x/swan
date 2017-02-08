package experiment

import (
	"os"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/swan/pkg/conf"
)

// MetadataConfig encodes the settings for connecting to the database.
type MetadataConfig struct {
	CassandraAddress           string
	CassandraPort              int
	CassandraUsername          string
	CassandraPassword          string
	CassandraConnectionTimeout time.Duration
	CassandraSslEnabled        bool
	CassandraSslHostValidation bool
	CassandraSslCAPath         string
	CassandraSslCertPath       string
	CassandraSslKeyPath        string
}

// DefaultMetadataConfig returns a setup which use a Cassandra cluster running on localhost
// without any authentication or encryption.
func DefaultMetadataConfig() MetadataConfig {
	return MetadataConfig{
		CassandraAddress:           "127.0.0.1",
		CassandraUsername:          "",
		CassandraPassword:          "",
		CassandraConnectionTimeout: 0,
		CassandraSslEnabled:        false,
		CassandraSslHostValidation: false,
		CassandraSslCAPath:         "",
		CassandraSslCertPath:       "",
		CassandraSslKeyPath:        "",
	}
}

// MetadataConfigFromFlags applies the Cassandra settings from the command line flags and
// environment variables.
func MetadataConfigFromFlags() MetadataConfig {
	return MetadataConfig{
		CassandraAddress:           conf.CassandraAddress.Value(),
		CassandraUsername:          conf.CassandraUsername.Value(),
		CassandraPassword:          conf.CassandraPassword.Value(),
		CassandraConnectionTimeout: conf.CassandraConnectionTimeout.Value(),
		CassandraSslEnabled:        conf.CassandraSslEnabled.Value(),
		CassandraSslHostValidation: conf.CassandraSslHostValidation.Value(),
		CassandraSslCAPath:         conf.CassandraSslCAPath.Value(),
		CassandraSslCertPath:       conf.CassandraSslCertPath.Value(),
		CassandraSslKeyPath:        conf.CassandraSslKeyPath.Value(),
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
// Connect() still needs to be called to get an active Cassandra session.
func NewMetadata(experimentID string, config MetadataConfig) *Metadata {
	return &Metadata{
		experimentID: experimentID,
		config:       config,
	}
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

// Connect creates a session to the Cassandra cluster. This function should only be called once.
func (m *Metadata) Connect() error {
	cluster := gocql.NewCluster(m.config.CassandraAddress)

	// TODO(niklas): make consistency configurable.
	cluster.Consistency = gocql.One

	cluster.ProtoVersion = 4
	cluster.Timeout = m.config.CassandraConnectionTimeout

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

	if err := session.Query("CREATE KEYSPACE IF NOT EXISTS swan WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};").Exec(); err != nil {
		return err
	}

	// NOTE: Schema consistency check is ignored by CREATE query. (https://git-wip-us.apache.org/repos/asf?p=cassandra.git;a=blob_plain;f=doc/native_protocol_v4.spec)
	// To ensure schema consistency we perform a simple SELECT query on 'system_schema.keyspaces'.
	// Consistency level is taken from 'cluster.Consistency' variable, it can also be defined for individual Query.
	if err = session.Query("SELECT * FROM system_schema.keyspaces;").Exec(); err != nil {
		return err
	}

	if err = session.Query("CREATE TABLE IF NOT EXISTS swan.metadata (experiment_id text, time timestamp, metadata map<text,text>, PRIMARY KEY ((experiment_id), time),) WITH CLUSTERING ORDER BY (time DESC);").Exec(); err != nil {
		return err
	}

	// NOTE: Same issue as above.
	if err = session.Query("SELECT * FROM system_schema.keyspaces;").Exec(); err != nil {
		return err
	}

	return nil
}

// Record stores a key and value and associates with the experiment id.
func (m *Metadata) Record(key string, value string) error {
	metadata := MetadataMap{}
	metadata[key] = value

	if err := m.session.Query(`INSERT INTO swan.metadata (experiment_id, time, metadata) VALUES (?, ?, ?)`,
		m.experimentID, time.Now(), metadata).Exec(); err != nil {
		return err
	}

	return nil
}

// RecordMap stores a key and value map and associates with the experiment id.
func (m *Metadata) RecordMap(metadata MetadataMap) error {
	if err := m.session.Query(`INSERT INTO swan.metadata (experiment_id, time, metadata) VALUES (?, ?, ?)`,
		m.experimentID, time.Now(), metadata).Exec(); err != nil {
		return err
	}

	return nil
}

// RecordEnv adds all OS Envrionment variables that starts with prefix 'prefix'
// in the metadata information
func (m *Metadata) RecordEnv(prefix string) error {
	metadata := MetadataMap{}
	// Store environment configuration.
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			fields := strings.SplitN(env, "=", 2)
			metadata[fields[0]] = fields[1]
		}
	}
	if err := m.session.Query(`INSERT INTO swan.metadata (experiment_id, time, metadata) VALUES (?, ?, ?)`,
		m.experimentID, time.Now(), metadata).Exec(); err != nil {
		return err
	}
	return nil
}

// Get retrieves all metadata maps from the database.
func (m *Metadata) Get() ([]MetadataMap, error) {
	var metadata MetadataMap

	out := []MetadataMap{}

	iter := m.session.Query(`SELECT metadata FROM swan.metadata WHERE experiment_id = ?`, m.experimentID).Iter()
	for iter.Scan(&metadata) {
		out = append(out, metadata)
	}
	if err := iter.Close(); err != nil {
		return []MetadataMap{}, err
	}

	return out, nil
}

// Clear deletes all metadata entries associated with the current experiment id.
func (m *Metadata) Clear() error {
	if err := m.session.Query(`DELETE FROM swan.metadata WHERE experiment_id = ?`,
		m.experimentID).Exec(); err != nil {
		return err
	}

	return nil
}
