package main

import (
	//"os"
	"os/user"
	//"path"
	"time"

	"github.com/Sirupsen/logrus"
	//"github.com/intelsdi-x/snap/scheduler/wmap"
	//"github.com/intelsdi-x/swan/pkg/cassandra"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	//"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
)

var (
	// Aggressors flag.
	aggressorsFlag = conf.NewSliceFlag(
		"aggr", "Aggressor to run experiment with. You can state as many as you want (--aggr=l1d --aggr=membw)")

	// Mutilate configuration.
	percentileFlag     = conf.NewStringFlag("percentile", "Tail latency Percentile", "99")
	mutilateMasterFlag = conf.NewIPFlag(
		"mutilate_master",
		"Mutilate master host for remote executor. In case of 0 agents being specified it runs in agentless mode.",
		"127.0.0.1")
	mutilateAgentsFlag = conf.NewSliceFlag(
		"mutilate_agent",
		"Mutilate agent hosts for remote executor. Can be specified many times for multiple agents setup.")

	// Snap path.
	//snapCassandraPluginPath = conf.NewFileFlag(
	//	"snap_cassandra_plugin_path",
	//	"Path to snap cassandra plugin.",
	//	path.Join(os.Getenv("GOPATH"), "bin", "snap-plugin-publisher-cassandra"))

	mutilateMasterFlagDefault = "local"
)

// newRemote is helper for creating remotes with default sshConfig.
func newRemote(ip string) executor.Executor {
	// TODO(bp): Have ability to choose user using parameter here.
	user, err := user.Current()
	errutil.Check(err)

	sshConfig, err := executor.NewSSHConfig(ip, executor.DefaultSSHPort, user)
	errutil.Check(err)

	return executor.NewRemote(sshConfig)
}

func prepareSnapSessionLauncher() snap.SessionLauncher {
	var mutilateSnapSession snap.SessionLauncher

	// NOTE: For debug it is convenient to disable snap for some experiment runs.
	if snap.SnapdHTTPEndpoint.Value() != "none" {

		// Create connection with Snap.
		logrus.Info("Connecting to Snapd on ", snap.SnapdHTTPEndpoint.Value())
		// TODO(bp): Make helper for passing host:port or only host option here.

		//snapConnection, err := client.New(
		//	fmt.Sprintf("http://%s:%s", snap.AddrFlag.Value(), snap.DefaultDaemonPort),
		//	"v1",
		//	true,
		//)
		//errutil.Check(err)

		// Load the snap cassandra publisher plugin if not yet loaded.
		// TODO(bp): Make helper for that.
		//logrus.Debug("Checking if publisher cassandra is loaded.")
		//plugins := snap.NewPlugins(snapConnection)
		//loaded, err := plugins.IsLoaded("publisher", "cassandra")
		//errutil.Check(err)
		//
		//if !loaded {
		//	pluginPath := snapCassandraPluginPath.Value()
		//	if _, err := os.Stat(pluginPath); err != nil && os.IsNotExist(err) {
		//		logrus.Error("Cannot find snap cassandra plugin at %q", pluginPath)
		//	}
		//	err = plugins.LoadPlugins([]string{pluginPath})
		//	errutil.Check(err)
		//}

		// Define publisher.
		//publisher := wmap.NewPublishNode("cassandra", 2)
		//if publisher == nil {
		//	logrus.Fatal("Failed to create Publish Node for cassandra")
		//}
		//
		//publisher.AddConfigItem("server", cassandra.AddrFlag.Value())

		// Initialize Mutilate Snap Session.
		//pp := path.Join(fs.GetSwanBuildPath(), "snap-plugin-collector-mutilate")
		//logrus.Info("new snap session with mutilate plugin path:", pp)
		//if _, err := os.Stat(pp); err != nil && os.IsNotExist(err) {
		//	logrus.Fatalf("snap-plugin-collector-mutilate not found at %q", pp)
		//}

		mutilateConfig := mutilatesession.DefaultConfig()
		mutilateSnapSession, err := mutilatesession.NewSessionLauncher(mutilateConfig)
		errutil.Check(err)
		return mutilateSnapSession
	} else {
		logrus.Warn("Warn: snap workflows disabled!")
	}
	return mutilateSnapSession
}

func isManualPolicy() bool {
	return hpSetsFlag.Value() != "" && beSetsFlag.Value() != ""
}

// Check README.md for details of this experiment.
func main() {
	// Setup conf.
	conf.SetAppName("memcached-sensitivity-profile")
	conf.SetHelp(`Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
It executes workloads and triggers gathering of certain metrics like latency (SLI) and the achieved number of Request per Second (QPS/RPS)`)

	// Parse CLI.
	errutil.Check(conf.ParseFlags())

	logrus.SetLevel(conf.LogLevel())

	// Validate environment.
	validateOS()

	// Isolation configuration method.
	// TODO: needs update for different isolation per cpu
	var hpIsolation, beIsolation, l1Isolation, llcIsolation isolation.Isolation
	var aggressorFactory sensitivity.AggressorFactory
	if isManualPolicy() {
		hpIsolation, beIsolation = manualPolicy()
		aggressorFactory = sensitivity.NewSingleIsolationAggressorFactory(beIsolation)
		defer beIsolation.Clean()
	} else {
		// NOTE: Temporary hack for having multiple isolations in Sensitivity Profile.
		hpIsolation, l1Isolation, llcIsolation = sensitivityProfileIsolationPolicy()
		aggressorFactory = sensitivity.NewMultiIsolationAggressorFactory(l1Isolation, llcIsolation)
		defer l1Isolation.Clean()
		defer llcIsolation.Clean()
	}
	defer hpIsolation.Clean()

	// Initialize Memcached Launcher.
	localForHP := executor.NewLocalIsolated(hpIsolation)
	memcachedConfig := memcached.DefaultMemcachedConfig()
	memcachedLauncher := memcached.New(localForHP, memcachedConfig)

	// Initialize Mutilate Load Generator.
	mutilateConfig := mutilate.DefaultMutilateConfig()
	mutilateConfig.MemcachedHost = memcachedConfig.IP
	mutilateConfig.MemcachedPort = memcachedConfig.Port
	mutilateConfig.LatencyPercentile = percentileFlag.Value()
	mutilateConfig.TuningTime = 1 * time.Second

	// Special case to have ability to use local executor for mutilate master load generator.
	// This is needed for docker testing.
	var masterLoadGeneratorExecutor executor.Executor
	masterLoadGeneratorExecutor = executor.NewLocal()
	if mutilateMasterFlag.Value() != mutilateMasterFlagDefault {
		masterLoadGeneratorExecutor = newRemote(mutilateMasterFlag.Value())
	}

	// Pack agents.
	agentsLoadGeneratorExecutors := []executor.Executor{}
	for _, agent := range mutilateAgentsFlag.Value() {
		agentsLoadGeneratorExecutors = append(agentsLoadGeneratorExecutors, newRemote(agent))
	}
	logrus.Debugf("Added %d mutilate agent(s) to mutilate cluster", len(agentsLoadGeneratorExecutors))

	// Validate mutilate cluster executors and their limit of
	// number of open file descriptors. Sane mutilate configuration requires
	// more than default (1024) for mutilate cluster.
	validateExecutorsNOFILELimit(
		append(agentsLoadGeneratorExecutors, masterLoadGeneratorExecutor),
	)

	mutilateLoadGenerator := mutilate.NewCluster(
		masterLoadGeneratorExecutor,
		agentsLoadGeneratorExecutors,
		mutilateConfig)

	// Initialize aggressors with BE isolation.
	aggressors := []sensitivity.LauncherSessionPair{}
	for _, aggressorName := range aggressorsFlag.Value() {
		aggressor, err := aggressorFactory.Create(aggressorName)
		errutil.Check(err)

		aggressors = append(aggressors, aggressor)
	}

	// Snap Session for mutilate.
	mutilateSnapSession := prepareSnapSessionLauncher()

	// Create Experiment configuration from Conf.
	sensitivityExperiment := sensitivity.NewExperiment(
		conf.AppName(),
		conf.LogLevel(),
		sensitivity.DefaultConfiguration(),
		sensitivity.NewLauncherWithoutSession(memcachedLauncher),
		sensitivity.NewMonitoredLoadGenerator(mutilateLoadGenerator, mutilateSnapSession),
		aggressors,
	)

	// Run Experiment.
	err := sensitivityExperiment.Run()
	errutil.Check(err)
}
