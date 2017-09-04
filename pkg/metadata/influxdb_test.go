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

package metadata

import (
	"fmt"
	"testing"

	"github.com/intelsdi-x/swan/pkg/conf"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInfluxDB(t *testing.T) {
	Convey("While using metadata package", t, func() {
		influxDefConf := DefaultInfluxDBConfig()
		Convey("InfluxDB default config shall have default settings", func() {
			So(influxDefConf.dbName, ShouldEqual, conf.InfluxDBMetaName.Value())
			So(influxDefConf.httpConfig.Addr, ShouldEqual, fmt.Sprintf("http://%s:%d", conf.InfluxDBAddress.Value(), conf.InfluxDBPort.Value()))
			So(influxDefConf.httpConfig.Username, ShouldEqual, conf.InfluxDBUsername.Value())
			So(influxDefConf.httpConfig.Password, ShouldEqual, conf.InfluxDBPassword.Value())
		})
	})
}
