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
	"testing"

	"github.com/intelsdi-x/swan/pkg/metadata"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCassandraDB(t *testing.T) {

	testMetadata := []map[string]string{
		{
			"E-key1": "E-val1",
			"E-key2": "E-val2",
			"E-key3": "E-val3",
		},
		{
			"F-key1": "F-val1",
			"F-key2": "F-val2",
			"F-key3": "F-val3",
		},
		{
			"N-key1": "N-val1",
			"N-key2": "N-val2",
			"N-key3": "N-val3",
		},
		{
			"P-key1": "P-val1",
			"P-key2": "P-val2",
			"P-key3": "P-val3",
		},
		{
			"X-key1": "X-val1",
			"X-key2": "X-val2",
			"X-key3": "X-val3",
		},
	}

	Convey("While using metadata package", t, func() {

		id := uuid.New()
		cassandraDefConf := metadata.DefaultCassandraConfig()
		Convey("NewCassandra shall return no error on default config", func() {
			cassandraMetadata, err := metadata.NewCassandra(id, cassandraDefConf)
			So(err, ShouldBeNil)

			Convey("Record metadata to DB shall succeed on all kinds", func() {
				err = cassandraMetadata.Record("keyE", "valueE", metadata.TypeEnviron)
				So(err, ShouldBeNil)

				retMap, err := cassandraMetadata.GetByKind(metadata.TypeEnviron)
				So(err, ShouldBeNil)
				So(len(retMap), ShouldEqual, 1)
				So(retMap["keyE"], ShouldEqual, "valueE")

				err = cassandraMetadata.Record("keyN", "valueN", metadata.TypeEmpty)
				So(err, ShouldBeNil)

				retMap, err = cassandraMetadata.GetByKind(metadata.TypeEmpty)
				So(err, ShouldBeNil)
				So(len(retMap), ShouldEqual, 1)
				So(retMap["keyN"], ShouldEqual, "valueN")

				err = cassandraMetadata.Record("keyF", "valueF", metadata.TypeFlags)
				So(err, ShouldBeNil)

				retMap, err = cassandraMetadata.GetByKind(metadata.TypeFlags)
				So(err, ShouldBeNil)
				So(len(retMap), ShouldEqual, 1)
				So(retMap["keyF"], ShouldEqual, "valueF")

				err = cassandraMetadata.Record("keyP", "valueP", metadata.TypePlatform)
				So(err, ShouldBeNil)

				retMap, err = cassandraMetadata.GetByKind(metadata.TypePlatform)
				So(err, ShouldBeNil)
				So(len(retMap), ShouldEqual, 1)
				So(retMap["keyP"], ShouldEqual, "valueP")

				err = cassandraMetadata.Record("keyX", "valueX", "metadata.TypeCustom")
				So(err, ShouldBeNil)

				retMap, err = cassandraMetadata.GetByKind("metadata.TypeCustom")
				So(err, ShouldBeNil)
				So(len(retMap), ShouldEqual, 1)
				So(retMap["keyX"], ShouldEqual, "valueX")

				retMap, err = cassandraMetadata.GetByKind("abcd")
				So(err, ShouldNotBeNil)
				So(len(retMap), ShouldEqual, 0)

				err = cassandraMetadata.Clear()
				So(err, ShouldBeNil)

				retMap, err = cassandraMetadata.GetByKind(metadata.TypePlatform)
				So(err, ShouldNotBeNil)
				So(len(retMap), ShouldEqual, 0)

			})

			Convey("RecordMap shall store all types of metadata", func() {
				err = cassandraMetadata.RecordMap(testMetadata[0], metadata.TypeEnviron)
				So(err, ShouldBeNil)
				retMap, err := cassandraMetadata.GetByKind(metadata.TypeEnviron)
				So(err, ShouldBeNil)
				So(len(retMap), ShouldEqual, len(testMetadata[0]))
				for key := range testMetadata[0] {
					So(testMetadata[0][key], ShouldEqual, retMap[key])
				}

				err = cassandraMetadata.RecordMap(testMetadata[1], metadata.TypeFlags)
				So(err, ShouldBeNil)
				retMap, err = cassandraMetadata.GetByKind(metadata.TypeFlags)
				So(err, ShouldBeNil)
				So(len(retMap), ShouldEqual, len(testMetadata[1]))
				for key := range testMetadata[1] {
					So(testMetadata[1][key], ShouldEqual, retMap[key])
				}

				err = cassandraMetadata.RecordMap(testMetadata[2], metadata.TypeEmpty)
				So(err, ShouldBeNil)
				retMap, err = cassandraMetadata.GetByKind(metadata.TypeEmpty)
				So(err, ShouldBeNil)
				So(len(retMap), ShouldEqual, len(testMetadata[2]))
				for key := range testMetadata[2] {
					So(testMetadata[2][key], ShouldEqual, retMap[key])
				}

				err = cassandraMetadata.RecordMap(testMetadata[3], metadata.TypePlatform)
				So(err, ShouldBeNil)
				retMap, err = cassandraMetadata.GetByKind(metadata.TypePlatform)
				So(err, ShouldBeNil)
				So(len(retMap), ShouldEqual, len(testMetadata[3]))
				for key := range testMetadata[3] {
					So(testMetadata[3][key], ShouldEqual, retMap[key])
				}

				err = cassandraMetadata.RecordMap(testMetadata[4], "metadata.TypeCustom")
				So(err, ShouldBeNil)
				retMap, err = cassandraMetadata.GetByKind("metadata.TypeCustom")
				So(err, ShouldBeNil)
				So(len(retMap), ShouldEqual, len(testMetadata[4]))
				for key := range testMetadata[4] {
					So(testMetadata[4][key], ShouldEqual, retMap[key])
				}

				err = cassandraMetadata.Clear()
				So(err, ShouldBeNil)

				retMap, err = cassandraMetadata.GetByKind(metadata.TypeEnviron)
				So(err, ShouldNotBeNil)
				So(len(retMap), ShouldEqual, 0)

			})

		})

	})
}
