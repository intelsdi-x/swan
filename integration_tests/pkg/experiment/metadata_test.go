package experiment

import (
	"testing"

	"github.com/intelsdi-x/swan/pkg/experiment"
	. "github.com/smartystreets/goconvey/convey"
)

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
				metadataCollection, err := metadata.Get()
				So(err, ShouldBeNil)
				So(metadataCollection, ShouldHaveLength, 2)

				// We don't test the contents here, as the order of the maps is not guaranteed.
				metadata.Clear()
			})
		})
	})
}
