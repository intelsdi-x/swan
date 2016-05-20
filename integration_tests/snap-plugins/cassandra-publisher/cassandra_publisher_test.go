package cassandra

import (
	"testing"
	"time"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	log "github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	"os"
	"path"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
)


func TestCassandraPublisher(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	snapd := testhelpers.NewSnapd()
	err := snapd.Start()
	Convey("Snapd should start successfully", t, func() {
		So(err, ShouldBeNil)
		So(snapd.Connected(), ShouldBeTrue)
	})
	defer snapd.Stop()
	defer snapd.CleanAndEraseOutput()

	snapdAddress := fmt.Sprintf("http://127.0.0.1:%d", snapd.Port())
	snapClient, err := client.New(snapdAddress, "v1", true)
	Convey("Snap client should connect successfully", t, func() {
		So(err, ShouldBeNil)
	})

	err = loadSnapPlugins(snapClient)
	Convey("Snap plugins loading is successfull", t, func() {
		So(err, ShouldBeNil)
	})

	err = runCassandraPublisherWorkflow(snapClient)
	Convey("Cassadndra workflow runs successfully", t, func() {
		So(err, ShouldBeNil)
	})

	value, tags, err := getValueAndTagsFromCassandra()
	Convey("When getting values from Cassadndra", t, func() {
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

func loadSnapPlugins(snapClient *client.Client) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic: %v\n", r)
			err = r.(error)
		}
	}()

	pluginClient := snap.NewPlugins(snapClient)
	requiredPlugins := getRequiredPlugins()

	for _, plugin := range requiredPlugins {
		if isNotPluginLoaded(pluginClient, plugin) {
			loadPlugin(pluginClient, plugin.pluginPath)
		}
	}

	return err
}

func isNotPluginLoaded(pluginClient *snap.Plugins, pi snapPluginInfo) (isLoaded bool){
	isLoaded, err := pluginClient.IsLoaded(pi.pluginType, pi.pluginName)
	if err != nil {
		panic(fmt.Errorf("Error while checking if plugin %s:%s is loaded: %s\n",
			pi.pluginType, pi.pluginName, err.Error()))
	}
	return !isLoaded
}

func loadPlugin(pluginClient *snap.Plugins, pluginPath string) {
	err := pluginClient.Load([]string{pluginPath})
	if err != nil {
		panic(fmt.Errorf("Could not load plugin in path: %s; %s\n",
			pluginPath, err.Error()))
	}
}

// Small struct for storing information for loading plugins.
type snapPluginInfo struct {
	pluginName string
	pluginType string
	pluginPath string
}

func getRequiredPlugins() (plugins []snapPluginInfo) {
	goPath := os.Getenv("GOPATH")
	plugins = make([]snapPluginInfo, 0, 2)
	plugins = append(plugins, snapPluginInfo{
		pluginName: "mock",
		pluginType: "collector",
		pluginPath: path.Join(
			goPath, "src", "github.com", "intelsdi-x", "swan",
			"build", "snap-plugin-collector-session-test"),
		})
	plugins = append(plugins, snapPluginInfo{
		pluginName: "cassandra",
		pluginType: "publisher",
		pluginPath: path.Join(goPath, "bin",
			"snap-plugin-publisher-cassandra"),
	})
	return plugins
}

func getValueAndTagsFromCassandra() (value float64, tags map[string]string, err error){
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

	session.Query(`SELECT doubleval, tags FROM snap.metrics
			WHERE ns = ? AND ver = ? AND host = ? LIMIT 1`,
				"/intel/swan/session/metric1", -1, "fedorowicz").
			Consistency(gocql.One).Scan(&value, &tags)
	return value, tags, err
}

func runCassandraPublisherWorkflow(snapClient *client.Client) (err error) {
	cassandraPublisher := wmap.NewPublishNode("cassandra", 2)
	cassandraPublisher.AddConfigItem("server", "localhost")

	snapSession := snap.NewSession(
		[]string{"/intel/swan/session/metric1"},
		1*time.Second,
		snapClient,
		cassandraPublisher)

	examplePhase := phase.Session{
		ExperimentID: "example-experiment",
		PhaseID     : "example-phase",
		RepetitionID: 42,
	}
	err = snapSession.Start(examplePhase)
	if err != nil {
		return fmt.Errorf("Snap session start failed: %s\n", err.Error())
	}

	time.Sleep(2 * time.Second)

	err = snapSession.Stop()
	if err != nil {
		return fmt.Errorf("Snap session stop failed: %s\n", err.Error())
	}

	return err
}
