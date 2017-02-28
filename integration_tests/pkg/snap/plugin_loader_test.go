package snap

import (
	"fmt"
	"testing"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/snap"
	. "github.com/smartystreets/goconvey/convey"
)

var plugins = []string{
	snap.DockerCollector,
	snap.MutilateCollector,
	snap.TagProcessor,
	snap.CassandraPublisher,
	snap.SessionPublisher,
}

func TestPluginLoader(t *testing.T) {

	Convey("While having Snapteld running", t, func() {
		cleanup, loader, snapteldAddr := testhelpers.RunAndTestSnaptel()
		defer cleanup()
		c, err := client.New(snapteldAddr, "v1", true)
		So(err, ShouldBeNil)

		pluginClient := snap.NewPlugins(c)
		for index, plugin := range plugins {
			Convey(fmt.Sprintf("We try to load %s plugin (%d)", plugin, index), func() {
				err := loader.Load(plugin)
				So(err, ShouldBeNil)
				Convey("Check if plugin is properly loaded", func() {
					pluginName, pluginType, err := snap.GetPluginNameAndType(plugin)
					So(err, ShouldBeNil)
					isLoaded, err := pluginClient.IsLoaded(pluginType, pluginName)
					So(err, ShouldBeNil)
					So(isLoaded, ShouldBeTrue)
				})
			})
		}
	})
}
