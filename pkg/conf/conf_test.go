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
	"io/ioutil"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConf(t *testing.T) {
	Convey("While using Conf pkg", t, func() {

		Convey("Log level can be fetched", func() {
			level, err := LogLevel()
			So(err, ShouldBeNil)
			So(level, ShouldEqual, logrus.InfoLevel)
		})

		Convey("Log level can be fetched from env", func() {
			level, err := LogLevel()
			So(err, ShouldBeNil)
			So(level, ShouldEqual, logrus.InfoLevel)

			os.Setenv(envName(logLevelFlag.Name), "debug")

			ParseFlags()

			// Should be from environment.
			level, err = LogLevel()
			So(err, ShouldBeNil)
			So(level, ShouldEqual, logrus.DebugLevel)
		})

		Convey("Validation for flags from env still works", func() {
			os.Setenv(envName(CassandraConnectionTimeout.Name), "foo-is-not-duration")
			err := ParseFlags()
			So(err, ShouldNotBeNil)
		})

		Convey("Validation for flags loaded from file", func() {
			const testfile = "testfile"
			err := ioutil.WriteFile(testfile, []byte("# comment\nfoo=baz"), os.ModePerm)
			So(err, ShouldBeNil)
			Reset(func() {
				os.Remove(testfile)
			})
			err = LoadConfig(testfile)
			So(err, ShouldBeNil)
			So(os.Getenv("SWAN_foo"), ShouldEqual, "baz")
		})
	})
}
