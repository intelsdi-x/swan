package cassandra

import (
	"testing"
	//"github.com/intelsdi-x/swan/pkg/executor"
	//"io/ioutil"
	"time"
	//"strings"
	//"regexp"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	log "github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	//"github.com/vektra/errors"
	"os"
	"path"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
)


func TestCassandraPublisher(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	snapClient, err := client.New("http://127.0.0.1:8181", "v1", true)
	Convey("When connecting to snap client", t, func() {
		So(err, ShouldBeNil)
	})


	err = loadSnapPlugins(snapClient)
	Convey("When loading snap plugins", t, func() {
		So(err, ShouldBeNil)
	})

	err = runCassandraPublisherWorkflow(snapClient)
	Convey("When running Cassadndra workflow", t, func() {
		So(err, ShouldBeNil)
	})

	value, tags, err := getValueAndTagsFromCassandra()
	Convey("When getting values from Cassadndra", t, func() {
		So(err, ShouldBeNil)
		Convey("Value should be different than zero", func() {
			So(value, ShouldNotEqual, 0)
		})
		Convey("Tags should be approprierate", func() {
			So(tags["swan_experiment"], ShouldEqual, "example-experiment")
			So(tags["swan_phase"], ShouldEqual, "example-phase")
			So(tags["swan_repetition"], ShouldEqual, "42")
		})
	})
}

type snapPluginInfo struct {
	pluginName string
	pluginType string
	pluginPath string
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

func getRequiredPlugins() (plugins []snapPluginInfo) {
	goPath := os.Getenv("GOPATH")
	plugins = make([]snapPluginInfo, 0, 2)
	plugins = append(plugins, snapPluginInfo{
		pluginName: "mock",
		pluginType: "collector",
		pluginPath: path.Join(goPath,
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
	//dummyMetricNS := "/intel/swan/session/metric1"
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
