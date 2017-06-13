// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cassandra

import (
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

	Convey("TestCassandraPublisher", t, func() {

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

		//cqlsh> select * from snap.metrics where ns='/intel/swan/session/metric1' AND ver=-1 AND host='fedorowicz' ORDER BY time ASC limit 1;
		// ns                          | ver | host       | time                            | boolval | doubleval | strval | tags                                                                                                                                 | valtype
		//-----------------------------+-----+------------+---------------------------------+---------+-----------+--------+--------------------------------------------------------------------------------------------------------------------------------------+-----------
		// /intel/swan/session/metric1 |  -1 | fedorowicz | 2016-05-20 11:07:02.890000+0000 |    null |         1 |   null | {'plugin_running_on': 'fedorowicz', 'swan_experiment': 'example-experiment', 'swan_phase': 'example-phase', 'swan_repetition': '42'} | doubleval
		valueFromMetrics, tagsFromMetrics, err := getMetricFromMetricsTable(`SELECT doubleval, tags FROM swan.metrics WHERE tags CONTAINS 'example-experiment' LIMIT 1 ALLOW FILTERING`)
		So(err, ShouldBeNil)

		//cqlsh> select * from snap.tags where key = 'swan_experiment' and val = 'example-experiment' ns='/intel/swan/session/metric1' AND ver=-1 AND host='fedorowicz' ORDER BY time ASC limit 1;
		// key                          | val                          | ns                          | ver | host       | time                            | boolval | doubleval | strval | tags                                                                                                                                 | valtype
		// -----------------------------+------------------------------+----------------+------------+-----+------------+---------------------------------+---------+------------+--------+-------------------------------------------------------------------------------------------------------------------------------------+-----------
		// swan_experiment              | example-experiment           | /intel/swan/session/metric1 |  -1 | fedorowicz | 2016-05-20 11:07:02.890000+0000 |    null |         1 |   null | {'plugin_running_on': 'fedorowicz', 'swan_experiment': 'example-experiment', 'swan_phase': 'example-phase', 'swan_repetition': '42'} | doubleval
		valueFromTags, tagsFromTags, err := getMetricFromMetricsTable(`SELECT doubleval, tags FROM swan.tags WHERE key = 'swan_experiment' AND val = 'example-experiment' LIMIT 1 ALLOW FILTERING`)
		So(err, ShouldBeNil)
		Convey("When getting values from Cassandra", func() {
			Convey("Values stored in Cassandra should be greater then 0", func() {
				So(valueFromMetrics, ShouldBeGreaterThan, 0)
				So(valueFromTags, ShouldBeGreaterThan, 0)
			})
			Convey("Tags should be appropriate", func() {
				So(tagsFromMetrics["swan_experiment"], ShouldEqual, "example-experiment")
				So(tagsFromTags["swan_experiment"], ShouldEqual, "example-experiment")
				So(tagsFromMetrics["swan_phase"], ShouldEqual, "example-phase")
				So(tagsFromTags["swan_phase"], ShouldEqual, "example-phase")
				So(tagsFromMetrics["swan_repetition"], ShouldEqual, "42")
				So(tagsFromTags["swan_repetition"], ShouldEqual, "42")
				So(tagsFromMetrics["FloatTag"], ShouldEqual, "42.123123")
				So(tagsFromTags["FloatTag"], ShouldEqual, "42.123123")
			})
		})
	})

}

func getMetricFromMetricsTable(cql string) (value float64, tags map[string]string, err error) {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.ProtoVersion = 4
	cluster.Keyspace = "swan"
	cluster.Consistency = gocql.All
	session, err := cluster.CreateSession()
	if err != nil {
		return 0, nil, errors.Wrapf(err, "cannot connect to cassandra: %v", cluster)
	}

	defer session.Close()

	// Try 5 times before giving up.
	for i := 0; i < 5; i++ {
		err = session.Query(cql).Consistency(gocql.One).Scan(&value, &tags)
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
	cassandraPublisher.AddConfigItem("server", "127.0.0.1")
	cassandraPublisher.AddConfigItem("keyspaceName", "swan")
	cassandraPublisher.AddConfigItem("tagIndex", "swan_experiment")

	snapSession := snap.NewSession(
		"swan-test-cassandra-publisher-session",
		[]string{"/intel/docker/root/stats/cgroups/cpu_stats/cpu_usage/total_usage"},
		1*time.Second,
		snapClient,
		cassandraPublisher)

	tags := make(map[string]interface{})
	tags[experiment.ExperimentKey] = "example-experiment"
	tags[experiment.PhaseKey] = "example-phase"
	tags[experiment.RepetitionKey] = 42
	tags["FloatTag"] = 42.123123
	handle, err := snapSession.Launch(tags)
	if err != nil {
		return errors.Errorf("snap session failed to start: %s", err.Error())
	}

	_, err = handle.Wait(0)
	if err != nil {
		return errors.Errorf("snap session wait failed: %s", err.Error())
	}

	err = handle.Stop()
	if err != nil {
		return errors.Errorf("snap session stop failed: %s", err.Error())
	}

	return nil
}
