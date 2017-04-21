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

package snap

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetPluginNameAndType(t *testing.T) {
	Convey("snap-plugin-publisher-cassandra should return type publisher and name cassandra", t, func() {
		name, pluginType, err := GetPluginNameAndType(CassandraPublisher)
		So(err, ShouldBeNil)
		So(name, ShouldEqual, "cassandra")
		So(pluginType, ShouldEqual, "publisher")
	})

	Convey("snap-plugin-collector-docker should return type collector and name docker", t, func() {
		name, pluginType, err := GetPluginNameAndType(DockerCollector)
		So(err, ShouldBeNil)
		So(name, ShouldEqual, "docker")
		So(pluginType, ShouldEqual, "collector")
	})

	Convey("snap-plugin-publisher-session-test should return type publisher and name session-test", t, func() {
		name, pluginType, err := GetPluginNameAndType(SessionPublisher)
		So(err, ShouldBeNil)
		So(name, ShouldEqual, "session-test")
		So(pluginType, ShouldEqual, "publisher")
	})
}
