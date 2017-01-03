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
