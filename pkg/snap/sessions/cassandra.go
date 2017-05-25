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

package sessions

import (
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/conf"
)

// ApplyCassandraConfiguration is a helper which applies the Cassandra related settings from
// the command line flags and applies them to a snap workflow.
func ApplyCassandraConfiguration(publisher *wmap.PublishWorkflowMapNode) {
	publisher.AddConfigItem("server", conf.CassandraAddress.Value())
	publisher.AddConfigItem("timeout", conf.CassandraConnectionTimeout.Value())
	publisher.AddConfigItem("username", conf.CassandraUsername.Value())
	publisher.AddConfigItem("password", conf.CassandraPassword.Value())
	publisher.AddConfigItem("ssl", conf.CassandraSslEnabled.Value())
	publisher.AddConfigItem("timeout", conf.CassandraTimeout.Value())
	publisher.AddConfigItem("connectionTimeout", conf.CassandraConnectionTimeout.Value())
	publisher.AddConfigItem("port", conf.CassandraPort.Value())
	publisher.AddConfigItem("initialHostLookup", conf.CassandraInitialHostLookup.Value())
	publisher.AddConfigItem("ignorePeerAddrRuleKey", conf.CassandraIgnorePeerAddr.Value())
	publisher.AddConfigItem("createKeyspace", conf.CassandraCreateKeyspace.Value())
	publisher.AddConfigItem("keyspaceName", conf.CassandraKeyspaceName.Value())

	if conf.CassandraSslEnabled.Value() {
		publisher.AddConfigItem("serverCertVerification", conf.CassandraSslHostValidation.Value())
		publisher.AddConfigItem("caPath", conf.CassandraSslCAPath.Value())
		publisher.AddConfigItem("certPath", conf.CassandraSslCertPath.Value())
		publisher.AddConfigItem("keyPath", conf.CassandraSslKeyPath.Value())
	}
}
