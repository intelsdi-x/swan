package conf

// CassandraAddress represents cassandra address flag.
var CassandraAddress = NewStringFlag("cassandra_addr", "Address of Cassandra DB endpoint", "127.0.0.1")
