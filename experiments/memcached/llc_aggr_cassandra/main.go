package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/cgroup"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
)

var (
	aggressorsFlag = conf.NewSliceFlag(
		"aggr", "Aggressor to run experiment with. "+
			"You can state as many as you want (--aggr=l1d --aggr=membw)")
	hpCPUCountFlag = conf.NewRegisteredIntFlag(
		"hp-cpus", "Number of CPUs assigned to high priority task", 4)
	beCPUCountFlag = conf.NewRegisteredIntFlag(
		"be-cpus", "Number of CPUs assigned to best effort task", 4)
)

// Check README.md for details of this experiment.
func main() {
	// Setup conf.
	conf.SetAppName("MemcachedWithMutilateToCassandra")
	conf.SetHelpPath(
		path.Join(fs.GetSwanExperimentPath(), "memcached", "llc_aggr_cassandra", "README.md"))

	// Parse CLI.
	err := conf.ParseFlagAndEnv()
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.SetLevel(conf.LogLevel())

	numaZero := isolation.NewIntSet(0)
	// Initialize Memcached Launcher with HP isolation.
	hpCpus, err := isolation.CPUSelect(hpCPUCountFlag.Value(), isolation.ShareLLCButNotL1L2, isolation.NewIntSet())
	if err != nil {
		logrus.Fatal(err)
	}

	hpIsolation, err := cgroup.NewCPUSet("hp", hpCpus, numaZero, true, false)
	if err != nil {
		logrus.Fatal(err)
	}

	err = hpIsolation.Create()
	if err != nil {
		logrus.Fatal(err)
	}

	defer hpIsolation.Clean()

	localForHP := executor.NewLocalIsolated(hpIsolation)
	memcachedLauncher := memcached.New(localForHP, memcached.DefaultMemcachedConfig())

	// Special case to have ability to use local executor for load generator.
	// This is needed for docker testing.
	var loadGeneratorExecutor executor.Executor
	loadGeneratorExecutor = executor.NewLocal()

	if workloads.LoadGeneratorAddrFlag.Value() != "local" {
		// Initialize Mutilate Launcher.
		user, err := user.Current()
		if err != nil {
			logrus.Fatal(err)
		}

		sshConfig, err := executor.NewSSHConfig(
			workloads.LoadGeneratorAddrFlag.Value(), executor.DefaultSSHPort, user)
		if err != nil {
			logrus.Fatal(err)
		}

		loadGeneratorExecutor = executor.NewRemote(sshConfig)
	}

	percentile, _ := decimal.NewFromString("99.9")
	mutilateConfig := mutilate.Config{
		MutilatePath:      mutilate.GetPathFromEnvOrDefault(),
		MemcachedHost:     "127.0.0.1",
		LatencyPercentile: percentile,
		TuningTime:        1 * time.Second,
	}
	mutilateLoadGenerator := mutilate.New(loadGeneratorExecutor, mutilateConfig)

	// Initialize BE isolation.
	beCpus, err := isolation.CPUSelect(beCPUCountFlag.Value(), isolation.ShareLLCButNotL1L2, hpCpus)
	if err != nil {
		logrus.Fatal(err)
	}

	beIsolation, err := cgroup.NewCPUSet("be", beCpus, numaZero, true, false)
	if err != nil {
		logrus.Fatal(err)
	}

	err = beIsolation.Create()
	if err != nil {
		logrus.Fatal(err)
	}

	defer beIsolation.Clean()

	// Initialize aggressors with BE isolation.
	aggressors := []sensitivity.LauncherSessionPair{}
	aggressorFactory := sensitivity.NewAggressorFactory(beIsolation)
	for _, aggr := range aggressorsFlag.Value() {
		aggressors = append(aggressors, aggressorFactory.Create(aggr))
	}

	// Create connection with Snap.
	logrus.Debug("Connecting to Snapd on ", snap.AddrFlag.Value())
	// TODO(bp): Make helper for passing host:port or only host option here.
	snapConnection, err := client.New(
		fmt.Sprintf("http://%s:%s", snap.AddrFlag.Value(), snap.DefaultDaemonPort),
		"v1",
		true,
	)
	if err != nil {
		logrus.Fatal(err)
	}

	// Load the snap cassandra publisher plugin if not yet loaded.
	// TODO(bp): Make helper for that.
	logrus.Debug("Checking if publisher cassandra is loaded.")
	plugins := snap.NewPlugins(snapConnection)
	loaded, err := plugins.IsLoaded("publisher", "cassandra")
	if err != nil {
		logrus.Fatal(err)
	}

	if !loaded {
		pluginPath := []string{path.Join(
			os.Getenv("GOPATH"), "bin", "snap-plugin-publisher-cassandra")}
		err = plugins.Load(pluginPath)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	// Define publisher.
	publisher := wmap.NewPublishNode("cassandra", 2)
	if publisher == nil {
		logrus.Fatal("Failed to create Publish Node for cassandra")
	}

	publisher.AddConfigItem("server", cassandra.AddrFlag.Value())

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
		aggressors,
	)

	// Run Experiment.
	err = sensitivityExperiment.Run()
	if err != nil {
		logrus.Fatal(err)
	}
}
