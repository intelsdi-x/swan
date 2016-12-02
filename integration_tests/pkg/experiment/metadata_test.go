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

		Convey("Recoding a metadata pair", func() {
			err := metadata.RecordMap(experiment.MetadataMap{"foo": "bar"})
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
	})
}
