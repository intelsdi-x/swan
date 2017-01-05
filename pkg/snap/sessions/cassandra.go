package sessions

import (
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

// ApplyCassandraConfiguration is a helper which applies the Cassandra related settings from
// the command line flags and applies them to a snap workflow.
func ApplyCassandraConfiguration(publisher *wmap.PublishWorkflowMapNode) {
	publisher.AddConfigItem("server", conf.CassandraAddress.Value())
	publisher.AddConfigItem("timeout", conf.CassandraConnectionTimeout.Value())
	publisher.AddConfigItem("username", conf.CassandraUsername.Value())
	publisher.AddConfigItem("password", conf.CassandraPassword.Value())
	publisher.AddConfigItem("ssl", conf.CassandraSslEnabled.Value())

	if conf.CassandraSslEnabled.Value() {
		publisher.AddConfigItem("serverCertVerification", conf.CassandraSslHostValidation.Value())
		publisher.AddConfigItem("caPath", conf.CassandraSslCAPath.Value())
		publisher.AddConfigItem("certPath", conf.CassandraSslCertPath.Value())
		publisher.AddConfigItem("keyPath", conf.CassandraSslKeyPath.Value())
	}
}
