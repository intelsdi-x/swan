package cassandra

import "github.com/gocql/gocql"

func configureCluster(ip string, keyspace string) *gocql.ClusterConfig {
	cluster := gocql.NewCluster(ip)
	cluster.Keyspace = keyspace
	cluster.ProtoVersion = 4
	cluster.Consistency = gocql.All
	return cluster
}

func CreateSession(ip string, keyspace string) (*gocql.Session, error) {
	cluster := configureCluster(ip, keyspace)
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}
