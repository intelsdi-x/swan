package conf

import (
	"time"
)

// CassandraAddress represents cassandra address flag.
var CassandraAddress = NewStringFlag("cassandra_addr", "Address of Cassandra DB endpoint", "127.0.0.1")

// CassandraUsername holds the user name which will be presented when connecting to the cluster at CassandraAddress.
var CassandraUsername = NewStringFlag("cassandra_username", "the user name which will be presented when connecting to the cluster", "")

// CassandraPassword holds the password which will be presented when connecting to the cluster at CassandraAddress.
var CassandraPassword = NewStringFlag("cassandra_password", "the password which will be presented when connecting to the cluster", "")

// CassandraConnectionTimeout encodes the internal connection timeout for the publisher. Note that increasing this
// value may increase the total connection time significantly, due to internal retry logic in the gocql library.
var CassandraConnectionTimeout = NewDurationFlag("cassandra_timeout", "the internal connection timeout for the publisher", 0*time.Second)

// CassandraSslEnabled determines whether the cassandra publisher should connect to the cluster over an SSL encrypted connection.
// Remember to set CassandraSslHostValidation, CassandraSslCAPath, CassandraSslCertPath and CassandraSslKeyPath accordingly.
var CassandraSslEnabled = NewBoolFlag("cassandra_ssl", "determines whether the cassandra publisher should connect to the cluster over an SSL encrypted connection", false)

// CassandraSslHostValidation determines whether the publisher will attempt to validate the cluster at CassandraAddress.
// Note that self-signed certificates and details like matching certificate hostname and the hostname connected to, will
// cause the connection to fail if not set up correctly.
// The recommended setting is to enable this flag.
var CassandraSslHostValidation = NewBoolFlag("cassandra_ssl_host_validation", "determines whether the publisher will attempt to validate the host", false)

// CassandraSslCAPath enables self-signed certificates by setting a certificate authority directly.
// This is not recommended in production settings.
var CassandraSslCAPath = NewStringFlag("cassandra_ssl_ca_path", "enables self-signed certificates by setting a certificate authority directly.", "")

// CassandraSslCertPath sets the client certificate, in case the cluster requires client verification.
var CassandraSslCertPath = NewStringFlag("cassandra_ssl_cert_path", "sets the client certificate, in case the cluster requires client verification", "")

// CassandraSslKeyPath sets the client private key, in case the cluster requires client verification.
var CassandraSslKeyPath = NewStringFlag("cassandra_ssl_key_path", "sets the client private key, in case the cluster requires client verification", "")
