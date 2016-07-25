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
	snap.KubesnapDockerCollector,
	snap.MutilateCollector,
	snap.CassandraPublisher,
	snap.SessionPublisher,
}

func TestPluginLoader(t *testing.T) {
	snapd := testhelpers.NewSnapd()
	err := snapd.Start()
	if err != nil {
		t.Fail()
	}
	defer snapd.Stop()
	defer snapd.CleanAndEraseOutput()

	snapdPort := snapd.Port()

	// Wait until snap is up.
	if !snapd.Connected() {
		t.Fail()
	}

	snapdAddress := fmt.Sprintf("http://%s:%d", "127.0.0.1", snapdPort)
	config := snap.DefaultPluginLoaderConfig()
	config.SnapdAddress = snapdAddress
	loader, err := snap.NewPluginLoader(config)
	if err != nil {
		t.Fail()
	}
	c, err := client.New(snapdAddress, "v1", true)
	pluginClient := snap.NewPlugins(c)

	Convey("While having Snapd running", t, func() {
		for index, plugin := range plugins {
			testPluginLoader(t, index, plugin, loader, pluginClient)
		}
	})
}

func testPluginLoader(t *testing.T, index int, pluginName string, loader *snap.PluginLoader, pluginClient *snap.Plugins) {
	Convey(fmt.Sprintf("We try to load %s plugin (%d)", pluginName, index), func() {
		err := loader.LoadPlugin(pluginName)
		So(err, ShouldBeNil)
		Convey("Check if plugin is properly loaded", func() {
			pluginName, pluginType := snap.GetPluginNameAndType(pluginName)
			isLoaded, err := pluginClient.IsLoaded(pluginType, pluginName)
			So(err, ShouldBeNil)
			So(isLoaded, ShouldBeTrue)
		})
	})
}
