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

package publishers

import (
	"fmt"

	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/snap"
)

// ApplyCassandraConfiguration is a helper which applies the Cassandra related settings from
// the command line flags and applies them to a snap workflow.
func ApplyInfluxDBConfiguration(publisher *wmap.PublishWorkflowMapNode) {
	fmt.Println("ApplyInfluxDBConfiguration called")
	publisher.AddConfigItem("host", conf.InfluxDBAddress.Value())
	publisher.AddConfigItem("user", conf.InfluxDBUsername.Value())
	publisher.AddConfigItem("password", conf.InfluxDBPassword.Value())
	publisher.AddConfigItem("database", conf.InfluxDBMetricsName.Value())
	publisher.AddConfigItem("port", conf.InfluxDBPort.Value())
	publisher.AddConfigItem("skip-verify", conf.InfluxDBInsecureSkipVerify.Value())
}

func NewDefaultInfluxDBPublisher() (pub Publisher) {
	pub.Publisher = wmap.NewPublishNode("influxdb", snap.PluginAnyVersion)
	ApplyInfluxDBConfiguration(pub.Publisher)

	pub.PluginName = snap.InfluxDBPublisher
	return
}
