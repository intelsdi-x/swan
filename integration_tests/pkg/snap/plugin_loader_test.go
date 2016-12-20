package snap

import (
	"fmt"
	"testing"

	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/utils/err_collection"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
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
	snapteld := testhelpers.NewSnapteld()
	err := snapteld.Start()
	if err != nil {
		t.Fatalf("snapteld creation failed: %q", err)
	}
	defer func() {
		var errCollection errcollection.ErrorCollection
		errCollection.Add(snapteld.Stop())
		errCollection.Add(snapteld.CleanAndEraseOutput())
		if err := errCollection.GetErrIfAny(); err != nil {
			t.Fatalf("Cleaning up procedures fails: %s", err)
		}
	}()

	snapteldAddress := fmt.Sprintf("http://%s:%d", "127.0.0.1", snapteld.Port())

	// Wait until snap is up.
	if !snapteld.Connected() {
		t.Fatalf("failed to connect to snapteld on %q", snapteldAddress)
	}

	config := snap.DefaultPluginLoaderConfig()
	config.SnapteldAddress = snapteldAddress
	loader, err := snap.NewPluginLoader(config)
	if err != nil {
		t.Fatalf("snap plugin loading failed: %q", err)
	}
	c, err := client.New(snapteldAddress, "v1", true)
	pluginClient := snap.NewPlugins(c)

	Convey("While having Snapteld running", t, func() {
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
