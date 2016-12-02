package experiment

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/athena/pkg/conf"
)

// MetadataConfig ...
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

// DefaultMetadataConfig ...
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

// MetadataConfigFromFlags ...
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

// MetadataMap ...
type MetadataMap map[string]string

// Metadata ...
type Metadata struct {
	experimentID string
	config       MetadataConfig
	session      *gocql.Session
}

// NewMetadata ...
func NewMetadata(experimentID string, config MetadataConfig) *Metadata {
	return &Metadata{
		experimentID: experimentID,
		config:       config,
	}
}

// Connect ...
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
		sslOptions := &gocql.SslOptions{
			EnableHostVerification: m.config.CassandraSslHostValidation,
		}

		if m.config.CassandraSslCAPath != "" {
			sslOptions.CaPath = m.config.CassandraSslCAPath
		}

		if m.config.CassandraSslCertPath != "" {
			sslOptions.CertPath = m.config.CassandraSslCertPath
		}

		if m.config.CassandraSslKeyPath != "" {
			sslOptions.KeyPath = m.config.CassandraSslKeyPath
		}

		cluster.SslOpts = sslOptions
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}

	m.session = session

	if err := session.Query("CREATE KEYSPACE IF NOT EXISTS swan WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};").Exec(); err != nil {
		return err
	}

	if err := session.Query("CREATE TABLE IF NOT EXISTS swan.metadata (experiment_id text, time timestamp, metadata map<text,text>, PRIMARY KEY ((experiment_id), time),) WITH CLUSTERING ORDER BY (time DESC);").Exec(); err != nil {
		return err
	}

	return nil
}

// Record ...
func (m *Metadata) Record(key string, value string) error {
	metadata := MetadataMap{}
	metadata[key] = value

	if err := m.session.Query(`INSERT INTO swan.metadata (experiment_id, time, metadata) VALUES (?, ?, ?)`,
		m.experimentID, time.Now(), metadata).Exec(); err != nil {
		return err
	}

	return nil
}

// RecordMap ...
func (m *Metadata) RecordMap(metadata MetadataMap) error {
	if err := m.session.Query(`INSERT INTO swan.metadata (experiment_id, time, metadata) VALUES (?, ?, ?)`,
		m.experimentID, time.Now(), metadata).Exec(); err != nil {
		return err
	}

	return nil
}

// Get ...
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

// Clear ...
func (m *Metadata) Clear() error {
	if err := m.session.Query(`DELETE FROM swan.metadata WHERE experiment_id = ?`,
		m.experimentID).Exec(); err != nil {
		return err
	}

	return nil
}
