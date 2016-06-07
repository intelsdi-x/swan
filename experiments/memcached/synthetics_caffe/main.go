package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/osutil"
	//"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/swan"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
	//"os"
	//"path"
	"time"
)

// Check README.md for details of this experiment.
func main() {
	logLevel := logrus.InfoLevel
	logrus.SetLevel(logLevel)

	hpIsolation, err := NewCgroup([]string{"cpuset"}, "hp")
	beIsolation, err := NewCgroup([]string{"cpuset"}, "be")

	//hpIsolation, err := NewCPUSetWithExecutor()

	local := executor.NewLocal()

	// Initialize Memcached Launcher.
	memcachedLauncher := memcached.New(local, memcached.DefaultMemcachedConfig())

	// Initialize Mutilate Launcher.
	percentile, _ := decimal.NewFromString("99.9")

	memcachedHost := osutil.GetEnvOrDefault("SWAN_MEMCAHED_HOST", "127.0.0.1")
	mutilateHost := osutil.GetEnvOrDefault("SWAN_MUTILATE_HOST", "127.0.0.1")
	mutilateConfig := mutilate.Config{
		MutilatePath:      mutilate.GetPathFromEnvOrDefault(),
		MemcachedHost:     memcachedHost,
		LatencyPercentile: percentile,
		TuningTime:        1 * time.Second,
	}

	mutilateLoadGenerator := mutilate.New(local, mutilateConfig)

	// Create connection with Snap.
	logrus.Debug("Connecting to Snap")

	snapAddress := osutil.GetEnvOrDefault("SWAN_SNAP_ADDRESS", "http://127.0.0.1:8181")
	snapConnection, err := client.New(snapAddress, "v1", true)
	if err != nil {
		panic(err)
	}

	// Load the snap cassandra publisher plugin if not yet loaded.
	// TODO(bp): Make helper for that.
	//logrus.Debug("Checking if publisher cassandra is loaded.")
	//plugins := snap.NewPlugins(snapConnection)
	//loaded, err := plugins.IsLoaded("publisher", "cassandra")
	//if err != nil {
	//	panic(err)
	//}

	//if !loaded {
	//	pluginPath := []string{path.Join(
	//		os.Getenv("GOPATH"), "bin", "snap-plugin-publisher-cassandra")}
	//	err = plugins.Load(pluginPath)
	//	if err != nil {
	//		panic(err)
	//	}
	//}

	// Define publisher.
	publisher := wmap.NewPublishNode("cassandra", 2)
	if publisher == nil {
		panic("Failed to create Publish Node for cassandra")
	}

	cassandraHostName := osutil.GetEnvOrDefault("SWAN_CASSANDRA_HOST", "127.0.0.1")
	publisher.AddConfigItem("server", cassandraHostName)

	// Initialize Mutilate Snap Session.
	mutilateSnapSession := sessions.NewMutilateSnapSessionLauncher(
		swan.GetSwanBuildPath(),
		1*time.Second,
		snapConnection,
		publisher)

	// Initialize aggressors.
	llcAggressorLauncher := l3data.New(local, l3data.DefaultL3Config())
	memBwAggressorLauncher := memoryBandwidth.New(local, memoryBandwidth.DefaultMemBwConfig())
	l1iAggressorLauncher := l1instruction.New(local, l1instruction.DefaultL1iConfig())
	lidAggressorLauncher := l1data.New(local, l1data.DefaultL1dConfig())
	caffeAggressorLauncher := caffe.New(local, caffe.DefaultConfig())

	// Create Experiment configuration.
	configuration := sensitivity.Configuration{
		SLO:             500,             // us
		LoadDuration:    1 * time.Second, //10 * time.Second,
		LoadPointsCount: 1,               //10,
		Repetitions:     1,               //3,
	}

	sensitivityExperiment := sensitivity.NewExperiment(
		"MemcachedWithLocalMutilateToCassandra",
		logLevel,
		configuration,
		sensitivity.NewLauncherWithoutSession(memcachedLauncher),
		sensitivity.NewMonitoredLoadGenerator(mutilateLoadGenerator, mutilateSnapSession),
		[]sensitivity.LauncherSessionPair{
			sensitivity.NewLauncherWithoutSession(l1iAggressorLauncher),
			sensitivity.NewLauncherWithoutSession(lidAggressorLauncher),
			sensitivity.NewLauncherWithoutSession(llcAggressorLauncher),
			sensitivity.NewLauncherWithoutSession(memBwAggressorLauncher),
			sensitivity.NewLauncherWithoutSession(caffeAggressorLauncher),
		},
	)

	// Run Experiment.
	err = sensitivityExperiment.Run()
	if err != nil {
		panic(err)
	}
}
