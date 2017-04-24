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

package conf

import (
	"time"
)

// CassandraAddress represents cassandra address flag.
var CassandraAddress = NewStringFlag("cassandra_address", "Address of Cassandra DB endpoint for Metadata and Snap Publishers.", "127.0.0.1")

// CassandraUsername holds the user name which will be presented when connecting to the cluster at CassandraAddress.
var CassandraUsername = NewStringFlag("cassandra_username", "The user name which will be presented when connecting to the cluster at 'cassandra_address'.", "")

// CassandraPassword holds the password which will be presented when connecting to the cluster at CassandraAddress.
var CassandraPassword = NewStringFlag("cassandra_password", "The password which will be presented when connecting to the cluster at 'cassandra_address'.", "")

// CassandraConnectionTimeout encodes the internal connection timeout for the publisher. Note that increasing this
// value may increase the total connection time significantly, due to internal retry logic in the gocql library.
var CassandraConnectionTimeout = NewDurationFlag("cassandra_timeout", "Timout for communication with Cassandra cluster.", 0*time.Second)

// CassandraSslEnabled determines whether the cassandra publisher should connect to the cluster over an SSL encrypted connection.
var CassandraSslEnabled = NewBoolFlag("cassandra_ssl", "Determines whether the cassandra publisher should connect to the cluster over an SSL encrypted connection. Flags CassandraSslHostValidation, CassandraSslCAPath, CassandraSslCertPath and CassandraSslKeyPath should be set accordingly.", false)

// CassandraSslHostValidation determines whether the publisher will attempt to validate the cluster at CassandraAddress.
var CassandraSslHostValidation = NewBoolFlag("cassandra_ssl_host_validation", "Determines whether the publisher will attempt to validate the host. Note that self-signed certificates and hostname mismatch, will cause the connection to fail if not set up correctly. The recommended setting is to enable this flag.", false)

// CassandraSslCAPath enables self-signed certificates by setting a certificate authority directly.
var CassandraSslCAPath = NewStringFlag("cassandra_ssl_ca_path", "Enables self-signed certificates by setting a certificate authority directly. This is not recommended in production settings.", "")

// CassandraSslCertPath sets the client certificate, in case the cluster requires client verification.
var CassandraSslCertPath = NewStringFlag("cassandra_ssl_cert_path", "Sets the client certificate, in case the cluster requires client verification.", "")

// CassandraSslKeyPath sets the client private key, in case the cluster requires client verification.
var CassandraSslKeyPath = NewStringFlag("cassandra_ssl_key_path", "Sets the client private key, in case the cluster requires client verification.", "")
