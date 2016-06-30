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
)

var (
	aggressorsFlag = conf.NewSliceFlag(
		"aggr", "Aggressor to run experiment with. You can state as many as you want (--aggr=l1d --aggr=membw)")
	hpCPUCountFlag = conf.NewIntFlag("hp_cpus", "Number of CPUs assigned to high priority task", 1)
	beCPUCountFlag = conf.NewIntFlag("be_cpus", "Number of CPUs assigned to best effort task", 1)
	hpCPUExclusive = conf.NewBoolFlag("hp_exclusive_cores", "Has high priority task exclusive cores", false)
	beCPUExclusive = conf.NewBoolFlag("be_exclusive_cores", "Has best effort task exclusive cores", false)
)

// ipAddressFlag returns IP which will be specified for workload services as endpoints.
var ipAddressFlag = conf.NewIPFlag(
	"ip",
	"IP of interface for Swan workloads services to listen on",
	"127.0.0.1",
)

// Check the supplied error, log and exit if non-nil.
func check(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}

// Check README.md for details of this experiment.
func main() {
	// Setup conf.
	conf.SetAppName("memcached-sensitivity-profile")
	conf.SetHelp(`Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
It executes workloads and triggers gathering of certain metrics like latency (SLI) and the achieved number of Request per Second (QPS/RPS)`)
	logrus.Info("1AAA")

	logrus.SetLevel(conf.LogLevel())
	logrus.Debug("1.2AAA")

	// Parse CLI.
	check(conf.ParseFlags())
	logrus.Debug("1.6AAA")

	threadSet := sharedCacheThreads()
	hpThreadIDs, err := threadSet.AvailableThreads().Take(hpCPUCountFlag.Value())
	check(err)

	// Allocate BE threads from the remaining threads on the same socket as the
	// HP workload.
	remaining := threadSet.AvailableThreads().Difference(hpThreadIDs)
	logrus.Debug("1.8AAA")
	beThreadIDs, err := remaining.Take(beCPUCountFlag.Value())
	check(err)
	logrus.Info("2AAA")

	// TODO(CD): Verify that it's safe to assume NUMA node 0 contains all
	// memory banks (probably not).
	numaZero := isolation.NewIntSet(0)

	// Initialize Memcached Launcher with HP isolation.
	hpIsolation, err := cgroup.NewCPUSet("hp", hpThreadIDs, numaZero, hpCPUExclusive.Value(), false)
	check(err)

	err = hpIsolation.Create()
	check(err)

	defer hpIsolation.Clean()

	localForHP := executor.NewLocalIsolated(hpIsolation)
	memcachedConfig := memcached.DefaultMemcachedConfig()
	memcachedConfig.IP = ipAddressFlag.Value()
	memcachedLauncher := memcached.New(localForHP, memcachedConfig)

	logrus.Info("3AAA")
	// Special case to have ability to use local executor for load generator.
	// This is needed for docker testing.
	var loadGeneratorExecutor executor.Executor
	loadGeneratorExecutor = executor.NewLocal()
	logrus.Info("4AAA")

	if workloads.LoadGeneratorAddrFlag.Value() != "local" {
		// Initialize Mutilate Launcher.
		user, err := user.Current()
		check(err)

		sshConfig, err := executor.NewSSHConfig(
			workloads.LoadGeneratorAddrFlag.Value(), executor.DefaultSSHPort, user)
		check(err)

		loadGeneratorExecutor = executor.NewRemote(sshConfig)
	}
	logrus.Info("5AAA")

	mutilateConfig := mutilate.DefaultMutilateConfig()
	mutilateConfig.MemcachedHost = memcachedConfig.IP
	mutilateConfig.MemcachedPort = memcachedConfig.Port
	mutilateConfig.LatencyPercentile = "99"
	mutilateConfig.TuningTime = 1 * time.Second
	logrus.Info("6AAA")

	mutilateLoadGenerator := mutilate.New(loadGeneratorExecutor, mutilateConfig)

	// Initialize BE isolation.
	beIsolation, err := cgroup.NewCPUSet("be", beThreadIDs, numaZero, beCPUExclusive.Value(), false)
	check(err)
	logrus.Info("7AAA")

	err = beIsolation.Create()
	check(err)

	defer beIsolation.Clean()

	// Initialize aggressors with BE isolation.
	aggressors := []sensitivity.LauncherSessionPair{}
	aggressorFactory := sensitivity.NewAggressorFactory(beIsolation)
	for _, aggr := range aggressorsFlag.Value() {
		aggressor, err := aggressorFactory.Create(aggr)
		if err != nil {
			logrus.Fatal(err)
		}
		aggressors = append(aggressors, aggressor)
	}
	logrus.Info("8AAA")

	// Create connection with Snap.
	logrus.Debug("Connecting to Snapd on ", snap.AddrFlag.Value())
	// TODO(bp): Make helper for passing host:port or only host option here.
	snapConnection, err := client.New(
		fmt.Sprintf("http://%s:%s", snap.AddrFlag.Value(), snap.DefaultDaemonPort),
		"v1",
		true,
	)
	check(err)

	// Load the snap cassandra publisher plugin if not yet loaded.
	// TODO(bp): Make helper for that.
	logrus.Debug("Checking if publisher cassandra is loaded.")
	plugins := snap.NewPlugins(snapConnection)
	loaded, err := plugins.IsLoaded("publisher", "cassandra")
	check(err)

	if !loaded {
		pluginPath := []string{path.Join(
			os.Getenv("GOPATH"), "bin", "snap-plugin-publisher-cassandra")}
		err = plugins.Load(pluginPath)
		check(err)
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
		LoadDuration:    5 * time.Second,
		LoadPointsCount: 1,
		Repetitions:     1,
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
	check(err)
}
