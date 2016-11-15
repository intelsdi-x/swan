package cassandra

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCassandraPublisher(t *testing.T) {

	snapd := testhelpers.NewSnapd()
	err := snapd.Start()
	if err != nil {
		t.Error(err)
	}
	defer snapd.Stop()
	defer snapd.CleanAndEraseOutput()

	if !snapd.Connected() {
		t.Error("Could not connect to snapd")
	}

	snapdAddress := fmt.Sprintf("http://127.0.0.1:%d", snapd.Port())
	snapClient, err := client.New(snapdAddress, "v1", true)
	if err != nil {
		t.Error(err)
	}

	err = loadSnapPlugins(snapdAddress)
	if err != nil {
		t.Error(err)
	}

	err = runCassandraPublisherWorkflow(snapClient)
	if err != nil {
		t.Error(err)
	}

	value, tags, err := getValueAndTagsFromCassandra()
	Convey("When getting values from Cassandra", t, func() {
		So(err, ShouldBeNil)
		Convey("Stored value in Cassandra should equal 1", func() {
			So(value, ShouldEqual, 1)
		})
		Convey("Tags should be approprierate", func() {
			So(tags["swan_experiment"], ShouldEqual, "example-experiment")
			So(tags["swan_phase"], ShouldEqual, "example-phase")
			So(tags["swan_repetition"], ShouldEqual, "42")
		})
	})
}

func loadSnapPlugins(snapdAddress string) (err error) {
	pluginLoaderConfig := snap.DefaultPluginLoaderConfig()
	pluginLoaderConfig.SnapdAddress = snapdAddress
	pluginLoader, err := snap.NewPluginLoader(pluginLoaderConfig)
	if err != nil {
		return err
	}

	return pluginLoader.Load(snap.CassandraPublisher, snap.SessionCollector)
}

func getValueAndTagsFromCassandra() (value float64, tags map[string]string, err error) {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.ProtoVersion = 4
	cluster.Keyspace = "snap"
	cluster.Consistency = gocql.All
	session, _ := cluster.CreateSession()
	defer session.Close()

	//cqlsh> select * from snap.metrics where ns='/intel/swan/session/metric1' AND ver=-1 AND host='fedorowicz' ORDER BY time ASC limit 1;
	//ns                          | ver | host       | time                            | boolval | doubleval | strval | tags                                                                                                                                 | valtype
	//-----------------------------+-----+------------+---------------------------------+---------+-----------+--------+--------------------------------------------------------------------------------------------------------------------------------------+-----------
	///intel/swan/session/metric1 |  -1 | fedorowicz | 2016-05-20 11:07:02.890000+0000 |    null |         1 |   null | {'plugin_running_on': 'fedorowicz', 'swan_experiment': 'example-experiment', 'swan_phase': 'example-phase', 'swan_repetition': '42'} | doubleval

	err = session.Query(`SELECT doubleval, tags FROM snap.metrics WHERE tags CONTAINS 'example-experiment' LIMIT 1 ALLOW FILTERING`).Consistency(gocql.One).Scan(&value, &tags)

	return value, tags, err
}

func runCassandraPublisherWorkflow(snapClient *client.Client) (err error) {
	cassandraName, _, err := snap.GetPluginNameAndType(snap.CassandraPublisher)
	if err != nil {
		return err
	}
	cassandraPublisher := wmap.NewPublishNode(cassandraName, snap.PluginAnyVersion)
	cassandraPublisher.AddConfigItem("server", "localhost")

	snapSession := snap.NewSession(
		[]string{"/intel/swan/session/metric1"},
		1*time.Second,
		snapClient,
		cassandraPublisher)

	tags := fmt.Sprintf("%s:%s,%s:%s,%s:%d",
		phase.ExperimentKey, "example-experiment",
		phase.PhaseKey, "example-phase",
		phase.RepetitionKey, 42)

	err = snapSession.Start(tags)
	if err != nil {
		return fmt.Errorf("snap session start failed: %s\n", err.Error())
	}

	snapSession.Wait()
	err = snapSession.Stop()
	if err != nil {
		return fmt.Errorf("snap session stop failed: %s\n", err.Error())
	}

	return err
}
