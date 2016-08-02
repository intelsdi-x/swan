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
	snap.SessionCollector,
	snap.TagProcessor,
	snap.CassandraPublisher,
	snap.SessionPublisher,
}

func TestPluginLoader(t *testing.T) {
	snapd := testhelpers.NewSnapd()
	err := snapd.Start()
	if err != nil {
		t.Fatalf("snapd creation failed: %q", err)
	}
	defer snapd.Stop()
	defer snapd.CleanAndEraseOutput()
	snapdAddress := fmt.Sprintf("http://%s:%d", "127.0.0.1", snapd.Port())

	// Wait until snap is up.
	if !snapd.Connected() {
		t.Fatalf("failed to connect to snapd on %q", snapdAddress)
	}

	config := snap.DefaultPluginLoaderConfig()
	config.SnapdAddress = snapdAddress
	loader, err := snap.NewPluginLoader(config)
	if err != nil {
		t.Fatalf("snap plugin loading failed: %q", err)
	}
	c, err := client.New(snapdAddress, "v1", true)
	pluginClient := snap.NewPlugins(c)

	Convey("While having Snapd running", t, func() {
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
