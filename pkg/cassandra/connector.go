package cassandra

import (
	"github.com/gocql/gocql"
	"github.com/vektra/errors"
)

func getClusterConfig(ip string, keyspace string) *gocql.ClusterConfig {
	cluster := gocql.NewCluster(ip)
	cluster.Keyspace = keyspace
	cluster.ProtoVersion = 4
	cluster.Consistency = gocql.All
	return cluster
}

// Connection has a session field which can be used to interact with cassandra.
type Connection struct {
	session *gocql.Session
}

func newConnection(session *gocql.Session) *Connection {
	return &Connection{session}
}

// CassandraSession returns current Cassandra session that can be used to iteract with indices.
func (cassandraConfig *Connection) CassandraSession() *gocql.Session {
	return cassandraConfig.session
}

// CreateConfigWithSession creates Cassandra config with prepared session.
func CreateConfigWithSession(ip string, keyspace string) (cassandraConfig *Connection, err error) {
	cluster := getClusterConfig(ip, keyspace)
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return newConnection(session), nil
}

// CloseSession closes current Cassandra session.
func (cassandraConfig *Connection) CloseSession() error {
	if !cassandraConfig.session.Closed() {
		cassandraConfig.session.Close()
		return nil
	}
	return errors.New("Session already closed")
}
