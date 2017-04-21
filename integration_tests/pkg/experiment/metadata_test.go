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

package experiment

import (
	"testing"

	"github.com/intelsdi-x/swan/pkg/experiment"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"time"
)

func TestMetadata(t *testing.T) {
	Convey("Connecting to the Cassandra database", t, func() {
		metadata, err := experiment.NewMetadata("foobar-experiment", experiment.DefaultMetadataConfig())
		So(err, ShouldBeNil)

		// If a test failed midway, there may be metadata associated with the 'foobar-experiment' above.
		metadata.Clear()

		// Make sure that metadata is cleared when test ends.
		Reset(func() {
			metadata.Clear()
			os.Unsetenv("SWAN_LOG")
		})

		Convey("Recording runtime environment", func() {
			os.Setenv("SWAN_LOG", "error")
			err := metadata.RecordRuntimeEnv(time.Now())
			So(err, ShouldBeNil)

			Convey("It should be possible to retrive environment variables, hostname, start time, flags and platform data", func() {
				metadataCollection, err := metadata.Get()
				So(err, ShouldBeNil)
				So(metadataCollection, ShouldHaveLength, 4)
				So(metadataCollection[0]["cpu_model"], ShouldNotBeNil)
				So(metadataCollection[1]["host"], ShouldNotBeNil)
				So(metadataCollection[1]["time"], ShouldNotBeNil)
				So(metadataCollection[2]["SWAN_LOG"], ShouldNotBeNil)
				So(metadataCollection[3]["cassandra_password"], ShouldNotBeNil)
			})
		})

		Convey("Recoding a metadata pair", func() {
			err := metadata.Record("foo", "bar")
			So(err, ShouldBeNil)

			Convey("Should be able to be fetched from Cassandra again", func() {
				metadataCollection, err := metadata.Get()
				So(err, ShouldBeNil)
				So(metadataCollection, ShouldHaveLength, 1)
				So(metadataCollection[0], ShouldContainKey, "foo")
				So(metadataCollection[0]["foo"], ShouldEqual, "bar")
				metadata.Clear()
			})
		})

		Convey("Recoding two metadata pairs", func() {
			err := metadata.RecordMap(experiment.MetadataMap{
				"foo": "bar",
				"bar": "baz",
			})
			So(err, ShouldBeNil)

			Convey("Should be able to be fetched from Cassandra again", func() {
				metadataCollection, err := metadata.Get()
				So(err, ShouldBeNil)
				So(metadataCollection, ShouldHaveLength, 1)
				So(metadataCollection[0], ShouldContainKey, "foo")
				So(metadataCollection[0]["foo"], ShouldEqual, "bar")
				So(metadataCollection[0], ShouldContainKey, "bar")
				So(metadataCollection[0]["bar"], ShouldEqual, "baz")
				metadata.Clear()
			})
		})

		Convey("Recoding metadata twice", func() {
			err := metadata.Record("foo", "bar")
			So(err, ShouldBeNil)

			err = metadata.Record("bar", "baz")
			So(err, ShouldBeNil)

			Convey("Should return two metadata maps", func() {
				metadataCollection, err := metadata.Get()
				So(err, ShouldBeNil)
				So(metadataCollection, ShouldHaveLength, 2)

				// We don't test the contents here, as the order of the maps is not guaranteed.
				metadata.Clear()
			})
		})
	})
}
