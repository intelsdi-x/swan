package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/utils"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
	"os"
	"path"
	"time"
)

// Check README.md for details of this experiment.
func main() {

	cli := utils.NewCliWithReadme(
		"MemcachedWithMutilateToCassandra",
		path.Join(utils.GetSwanExperimentPath(), "memcached", "llc_aggr_cassandra", "README.md")).
		AddCassandraHostArg().
		AddLoadGeneratorHostArg().
		AddSnapHostArg().
		MustParse()

	logrus.SetLevel(cli.LogLevel())
	local := executor.NewLocal()

	// Initialize Memcached Launcher.
	memcachedLauncher := memcached.New(local, memcached.DefaultMemcachedConfig())

	// Initialize Mutilate Launcher.
	percentile, _ := decimal.NewFromString("99.9")
	localIP, err := utils.GetLocalIPAddr()
	if err != nil {
		panic(err)
	}

	mutilateConfig := mutilate.Config{
		MutilatePath:      mutilate.GetPathFromEnvOrDefault(),
		MemcachedHost:     localIP.String(),
		LatencyPercentile: percentile,
		TuningTime:        1 * time.Second,
	}

	remote, err := executor.NewRemoteWithDefaultConfig(cli.Get(utils.LoadGeneratorHostArg))
	if err != nil {
		panic(err)
	}
	mutilateLoadGenerator := mutilate.New(remote, mutilateConfig)

	// Create connection with Snap.
	logrus.Debug("Connecting to Snap")
	snapConnection, err :=
		client.New(fmt.Sprintf("http://%s:8181", cli.Get(utils.SnapdHostArg)), "v1", true)
	if err != nil {
		panic(err)
	}

	// Load the snap cassandra publisher plugin if not yet loaded.
	// TODO(bp): Make helper for that.
	logrus.Debug("Checking if publisher cassandra is loaded.")
	plugins := snap.NewPlugins(snapConnection)
	loaded, err := plugins.IsLoaded("publisher", "cassandra")
	if err != nil {
		panic(err)
	}

	if !loaded {
		pluginPath := []string{path.Join(
			os.Getenv("GOPATH"), "bin", "snap-plugin-publisher-cassandra")}
		err = plugins.Load(pluginPath)
		if err != nil {
			panic(err)
		}
	}

	// Define publisher.
	publisher := wmap.NewPublishNode("cassandra", 2)
	if publisher == nil {
		panic("Failed to create Publish Node for cassandra")
	}

	publisher.AddConfigItem("server", cli.Get(utils.CassandraHostArg))

	// Initialize Mutilate Snap Session.
	mutilateSnapSession := sessions.NewMutilateSnapSessionLauncher(
		utils.GetSwanBuildPath(),
		1*time.Second,
		snapConnection,
		publisher)

	// Initialize LLC aggressor.
	llcAggressorLauncher := l3data.New(local, l3data.DefaultL3Config())

	// Create Experiment configuration.
	configuration := sensitivity.Configuration{
		SLO:             500,             // us
		LoadDuration:    1 * time.Second, //10 * time.Second,
		LoadPointsCount: 1,               //10,
		Repetitions:     1,               //3,
	}

	sensitivityExperiment := sensitivity.NewExperiment(
		cli.AppName,
		cli.LogLevel(),
		configuration,
		sensitivity.NewLauncherWithoutSession(memcachedLauncher),
		sensitivity.NewMonitoredLoadGenerator(mutilateLoadGenerator, mutilateSnapSession),
		[]sensitivity.LauncherSessionPair{
			sensitivity.NewLauncherWithoutSession(llcAggressorLauncher),
		},
	)

	// Run Experiment.
	err = sensitivityExperiment.Run()
	if err != nil {
		panic(err)
	}
}
