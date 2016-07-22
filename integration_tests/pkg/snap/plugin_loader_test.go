package snap

import (
	"fmt"
	"testing"

	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/snap"
	. "github.com/smartystreets/goconvey/convey"
)

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

	config := snap.DefaultPluginLoaderConfig()
	config.SnapdAddress = fmt.Sprintf("http://%s:%d", "127.0.0.1", snapdPort)
	loader, err := snap.NewPluginLoader(config)
	if err != nil {
		t.Fail()
	}

	Convey("While having Snapd running", t, func() {
		Convey("We try to load Kubesnap Docker Publisher plugin", func() {
			err = loader.LoadPlugin(snap.KubesnapDockerCollector)
			So(err, ShouldBeNil)
		})

		Convey("We try to load Mutilate Collector plugin", func() {
			err = loader.LoadPlugin(snap.MutilateCollector)
			So(err, ShouldBeNil)
		})

		Convey("We try to load Cassandra Publisher plugin", func() {
			err = loader.LoadPlugin(snap.CassandraPublisher)
			So(err, ShouldBeNil)
		})

		Convey("We try to load Session Publisher plugin", func() {
			err = loader.LoadPlugin(snap.SessionPublisher)
			So(err, ShouldBeNil)
		})

		Convey("We try to load Cassandra Publisher twice", func() {
			err = loader.LoadPlugin(snap.CassandraPublisher)
			So(err, ShouldBeNil)
			err = loader.LoadPlugin(snap.CassandraPublisher)
			So(err, ShouldBeNil)
		})
	})
}
