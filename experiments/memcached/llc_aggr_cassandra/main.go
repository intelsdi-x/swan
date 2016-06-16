package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
	"os"
	"os/user"
	"path"
	"time"
)

// Check README.md for details of this experiment.
func main() {
	// Setup conf.
	conf.SetAppName("MemcachedWithLocalMutilateToCassandra")
	conf.SetHelpPath(
		path.Join(fs.GetSwanExperimentPath(), "memcached", "llc_aggr_cassandra", "README.md"))

	logrus.SetLevel(conf.LogLevel())

	memcachedConfig := memcached.DefaultMemcachedConfig()

	percentile, _ := decimal.NewFromString("99.9")
	mutilateConfig := mutilate.Config{
		MutilatePath:      mutilate.GetPathFromEnvOrDefault(),
		MemcachedHost:     "127.0.0.1",
		LatencyPercentile: percentile,
		TuningTime:        1 * time.Second,
	}

	l3dataConfig := l3data.DefaultL3Config()

	snapdAddr := snap.AddrFlag()
	cassandraAddr := cassandra.AddrFlag()
	loadGeneratorHost := workloads.LoadGeneratorAddrFlag()

	// Parse CLI.
	err := conf.ParseFlagAndEnv()
	if err != nil {
		panic(err)
	}

	// Initialize Memcached Launcher.
	local := executor.NewLocal()
	memcachedLauncher := memcached.New(local, memcachedConfig)

	// Special case to have ability to use local executor for load generator.
	// This is needed for docker testing.
	var loadGeneratorExecutor executor.Executor
	loadGeneratorExecutor = local

	if *loadGeneratorHost != "local" {
		// Initialize Mutilate Launcher.
		user, err := user.Current()
		if err != nil {
			panic(err)
		}

		sshConfig, err := executor.NewSSHConfig(*loadGeneratorHost, executor.DefaultSSHPort, user)
		if err != nil {
			panic(err)
		}

		loadGeneratorExecutor = executor.NewRemote(sshConfig)
	}

	mutilateLoadGenerator := mutilate.New(loadGeneratorExecutor, mutilateConfig)

	// Initialize LLC aggressor.
	llcAggressorLauncher := l3data.New(local, l3dataConfig)

	// Create connection with Snap.
	logrus.Debug("Connecting to Snapd on ", *snapdAddr)
	// TODO(bp): Make helper for passing host:port or only host option here.
	snapConnection, err := client.New(
		fmt.Sprintf("http://%s:%s", *snapdAddr, snap.DefaultDaemonPort),
		"v1",
		true,
	)

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

	publisher.AddConfigItem("server", *cassandraAddr)

	// Initialize Mutilate Snap Session.
	mutilateSnapSession := sessions.NewMutilateSnapSessionLauncher(
		fs.GetSwanBuildPath(),
		1*time.Second,
		snapConnection,
		publisher)

	// Create Experiment configuration.
	configuration := sensitivity.Configuration{
		SLO:             500, // us
		LoadDuration:    10 * time.Second,
		LoadPointsCount: 10,
		Repetitions:     3,
	}

	sensitivityExperiment := sensitivity.NewExperiment(
		conf.AppName(),
		conf.LogLevel(),
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
