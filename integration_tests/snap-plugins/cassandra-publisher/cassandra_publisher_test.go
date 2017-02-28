package cassandra

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCassandraPublisher(t *testing.T) {

	Convey("description", t, func() {

		cleanup, loader, snapteldAddr := testhelpers.RunAndTestSnaptel()
		defer cleanup()

		err := loader.Load(snap.CassandraPublisher, snap.DockerCollector)
		So(err, ShouldBeNil)

		snapClient, err := client.New(snapteldAddr, "v1", true)
		So(err, ShouldBeNil)

		err = runCassandraPublisherWorkflow(snapClient)
		So(err, ShouldBeNil)

		// Given enough time before checking data was stored - replace with onetime tasks in snap when available.
		time.Sleep(5 * time.Second)

		value, tags, err := getValueAndTagsFromCassandra()
		Convey("When getting values from Cassandra", t, func() {
			So(err, ShouldBeNil)
			Convey("Stored value in Cassandra should be greater then 0", func() {
				So(value, ShouldBeGreaterThan, 0)
			})
			Convey("Tags should be approprierate", func() {
				So(tags["swan_experiment"], ShouldEqual, "example-experiment")
				So(tags["swan_phase"], ShouldEqual, "example-phase")
				So(tags["swan_repetition"], ShouldEqual, "42")
			})
		})
	})

}

func getValueAndTagsFromCassandra() (value float64, tags map[string]string, err error) {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.ProtoVersion = 4
	cluster.Keyspace = "snap"
	cluster.Consistency = gocql.All
	session, err := cluster.CreateSession()
	if err != nil {
		return 0, nil, errors.Wrapf(err, "cannot connect to cassandra: %v", cluster)
	}

	defer session.Close()

	//cqlsh> select * from snap.metrics where ns='/intel/swan/session/metric1' AND ver=-1 AND host='fedorowicz' ORDER BY time ASC limit 1;
	//ns                          | ver | host       | time                            | boolval | doubleval | strval | tags                                                                                                                                 | valtype
	//-----------------------------+-----+------------+---------------------------------+---------+-----------+--------+--------------------------------------------------------------------------------------------------------------------------------------+-----------
	///intel/swan/session/metric1 |  -1 | fedorowicz | 2016-05-20 11:07:02.890000+0000 |    null |         1 |   null | {'plugin_running_on': 'fedorowicz', 'swan_experiment': 'example-experiment', 'swan_phase': 'example-phase', 'swan_repetition': '42'} | doubleval

	// Try 5 times before giving up.
	for i := 0; i < 5; i++ {
		err = session.Query(`SELECT doubleval, tags FROM snap.metrics WHERE tags CONTAINS 'example-experiment' LIMIT 1 ALLOW FILTERING`).Consistency(gocql.One).Scan(&value, &tags)
		if err == nil || err != gocql.ErrNotFound {
			break
		}
		time.Sleep(1 * time.Second)
	}

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
		"swan-test-cassandra-publisher-session",
		[]string{"/intel/docker/root/stats/cgroups/cpu_stats/cpu_usage/total_usage"},
		1*time.Second,
		snapClient,
		cassandraPublisher)

	tags := fmt.Sprintf("%s:%s,%s:%s,%s:%d",
		experiment.ExperimentKey, "example-experiment",
		experiment.PhaseKey, "example-phase",
		experiment.RepetitionKey, 42)

	err = snapSession.Start(tags)
	if err != nil {
		return fmt.Errorf("snap session start failed: %s", err.Error())
	}

	snapSession.Wait()
	err = snapSession.Stop()
	if err != nil {
		return fmt.Errorf("snap session stop failed: %s", err.Error())
	}

	return err
}
