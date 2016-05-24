package cassandra

import (
	"github.com/gocql/gocql"
)

func getClusterConfig(ip string, keyspace string) *gocql.ClusterConfig {
	cluster := gocql.NewCluster(ip)
	cluster.Keyspace = keyspace
	cluster.ProtoVersion = 4
	cluster.Consistency = gocql.All
	return cluster
}

// CreateConfigWithSession creates Cassandra config with prepared session.
func CreateConfigWithSession(ip string, keyspace string) (cassandraConfig *Config, err error) {
	cluster := getClusterConfig(ip, keyspace)
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return newConfig(session), nil
}

// Config has a session field which can be used to interact with cassandra.
type Config struct {
	session *gocql.Session
}

func newConfig(session *gocql.Session) *Config {
	return &Config{session}
}

// CassandraSession returns current Cassandra session that can be used to iteract with indices.
func (cassandraConfig *Config) CassandraSession() *gocql.Session {
	return cassandraConfig.session
}
