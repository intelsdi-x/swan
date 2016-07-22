package snap

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetPluginNameAndType(t *testing.T) {
	Convey("snap-plugin-publisher-cassandra should return type publisher and name cassandra", t, func() {
		name, pluginType := GetPluginNameAndType(CassandraPublisher)
		So(name, ShouldEqual, "cassandra")
		So(pluginType, ShouldEqual, "publisher")
	})

	Convey("kubesnap-plugin-collector-docker should return type collector and name docker", t, func() {
		name, pluginType := GetPluginNameAndType(KubesnapDockerCollector)
		So(name, ShouldEqual, "docker")
		So(pluginType, ShouldEqual, "collector")
	})
}
