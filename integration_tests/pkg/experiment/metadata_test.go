package experiment

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/intelsdi-x/swan/pkg/experiment"
	. "github.com/smartystreets/goconvey/convey"
)

func dump(i interface{}) string {
	s, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(s)
}

func TestMetadata(t *testing.T) {
	metadata := experiment.NewMetadata("foobar-experiment", experiment.DefaultMetadataConfig())
	Convey("Connecting to the Cassandra database", t, func() {
		err := metadata.Connect()
		So(err, ShouldBeNil)

		// If a test failed midway, there may be metadata associated with the 'foobar-experiment' above.
		metadata.Clear()

		// Make sure that metadata is cleared when test ends.
		Reset(func() {
			metadata.Clear()
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

				// Cassandra doesn't support read-your-own writes consistency - try 10s then give up.
				// https://issues.apache.org/jira/browse/CASSANDRA-876
				N := 10
				found := false
				for i := 0; i < N; i++ {

					metadataCollection, err := metadata.Get()
					So(err, ShouldBeNil)

					if len(metadataCollection) == 2 {
						found = true
					}
					time.Sleep(1 * time.Second)
				}
				So(found, ShouldBeTrue)
				metadata.Clear()
			})
		})

		Convey("Recoding metadata map as group", func() {

			testData := experiment.MetadataMap{"foo": "bar"}

			_, err := metadata.GetGroup("testgroup")
			So(err, ShouldNotBeNil)

			err = metadata.RecordMapGroup(testData, "testgroup")
			So(err, ShouldBeNil)

			Convey("Should return one metadata group", func() {
				metadataGroup, err := metadata.GetGroup("testgroup")
				So(err, ShouldBeNil)
				So(metadataGroup, ShouldResemble, testData)

				// We don't test the contents here, as the order of the maps is not guaranteed.
				metadata.Clear()
			})
		})
	})
}
