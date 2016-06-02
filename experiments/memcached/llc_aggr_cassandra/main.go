package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/utils"
	"github.com/intelsdi-x/swan/pkg/workloads"
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
	// CLI argument registration.
	cli := utils.NewCliWithReadme(
		"MemcachedWithMutilateToCassandra",
		path.Join(utils.GetSwanExperimentPath(), "memcached", "llc_aggr_cassandra", "README.md"))
	cassandraAddr := cli.RegisterStringArgWithEnv(cassandra.AddressArg())
	lgAddr := cli.RegisterStringArgWithEnv(workloads.LoadGeneratorAddressArg())
	snapAddr := cli.RegisterStringArgWithEnv(snap.AddressArg())
	// Binaries paths.
	memcachedPath := cli.RegisterStringArgWithEnv(memcached.PathArg())
	mutilatePath := cli.RegisterStringArgWithEnv(mutilate.PathArg())
	l3Path := cli.RegisterStringArgWithEnv(l3data.PathArg())

	cli.MustParse()

	logrus.SetLevel(cli.LogLevel())

	// Initialize Memcached Launcher.
	local := executor.NewLocal()
	memcachedConfig := memcached.DefaultMemcachedConfig()
	memcachedConfig.PathToBinary = *memcachedPath
	memcachedConfig.ServerIP = cli.IPAddress()
	memcachedLauncher := memcached.New(local, memcachedConfig)

	// Initialize Mutilate Launcher.
	percentile, _ := decimal.NewFromString("99.9")

	mutilateConfig := mutilate.Config{
		MutilatePath:      *mutilatePath,
		MemcachedHost:     cli.IPAddress(),
		LatencyPercentile: percentile,
		TuningTime:        1 * time.Second,
	}

	var lgExecutor executor.Executor
	var err error
	// NOTE: We don't want to ssh on localhost if not needed - this enables ease of use inside
	// docker with net=host flag.
	if utils.IsAddrLocal(*lgAddr) {
		lgExecutor = local
	} else {
		lgExecutor, err = executor.NewRemoteWithDefaultConfig(*lgAddr)
		if err != nil {
			panic(err)
		}
	}

	mutilateLoadGenerator := mutilate.New(lgExecutor, mutilateConfig)

	// Create connection with Snap.
	// TODO(bp): Make Snap connection arg able to be specified as <host:port> or <host> and
	// default port will be added.
	logrus.Debug("Connecting to Snap")
	snapConnection, err :=
		client.New(fmt.Sprintf("http://%s:8181", *snapAddr), "v1", true)
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
		utils.GetSwanBuildPath(),
		1*time.Second,
		snapConnection,
		publisher)

	// Initialize LLC aggressor.
	l3Configuration := l3data.DefaultL3Config()
	l3Configuration.Path = *l3Path
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
