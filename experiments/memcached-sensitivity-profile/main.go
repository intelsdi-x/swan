package main

import (
	"os/user"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/isolation"
	"github.com/intelsdi-x/athena/pkg/kubernetes"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/athena/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
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
	runOnKubernetesFlag = conf.NewBoolFlag("run_on_kubernetes", "Launch HP and BE tasks on Kubernetes.", false)

	// Should CPUs be used exclusively?
	hpCPUExclusiveFlag = conf.NewBoolFlag("hp_exclusive_cores", "Has high priority task exclusive cores", false)
	beCPUExclusiveFlag = conf.NewBoolFlag("be_exclusive_cores", "Has best effort task exclusive cores", false)

	// For CPU count based isolation policy flags.
	hpCPUCountFlag = conf.NewIntFlag("hp_cpus", "Number of CPUs assigned to high priority task", 1)
	beCPUCountFlag = conf.NewIntFlag("be_cpus", "Number of CPUs assigned to best effort task", 1)

	// For manually provided isolation policy.
	hpSetsFlag = conf.NewStringFlag("hp_sets", "HP cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2", "")
	beSetsFlag = conf.NewStringFlag("be_sets", "BE cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2", "")

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

		mutilateConfig := mutilatesession.DefaultConfig()
		mutilateConfig.SnapdAddress = snap.SnapdHTTPEndpoint.Value()
		mutilateSnapSession, err := mutilatesession.NewSessionLauncher(mutilateConfig)
		errutil.Check(err)
		return mutilateSnapSession
	}
	return mutilateSnapSession
}

func isManualPolicy() bool {
	return hpSetsFlag.Value() != "" && beSetsFlag.Value() != ""
}

// DecoratorFunc is a dummy decorator as a workaround for a isolation for inside a docker.
type DecoratorFunc func(string) string

// Decorate wrap method for a command.
func (df DecoratorFunc) Decorate(command string) string {
	return df(command)
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
	var hpIsolation, beIsolation, l1Isolation, llcIsolation isolation.Decorator
	var aggressorFactory sensitivity.AggressorFactory

	if isManualPolicy() {
		manualTopology := newManualTopology(hpSetsFlag.Value(), beSetsFlag.Value(), hpCPUExclusiveFlag.Value(), beCPUExclusiveFlag.Value())
		hpIsolation = isolation.Numactl{PhyscpubindCPUs: manualTopology.hpCPUs, PreferredNode: manualTopology.hpNumaNodes[0]}
		beIsolation = isolation.Numactl{PhyscpubindCPUs: manualTopology.beCPUs, PreferredNode: manualTopology.beNumaNodes[0]}
		aggressorFactory = sensitivity.NewSingleIsolationAggressorFactory(beIsolation)
	} else {
		defaultTopology := newDefaultTopology(hpCPUCountFlag.Value(), beCPUCountFlag.Value(), hpCPUExclusiveFlag.Value(), beCPUExclusiveFlag.Value())
		hpIsolation = isolation.Numactl{PhyscpubindCPUs: defaultTopology.hpThreadIDs.AsSlice(), PreferredNode: defaultTopology.numaNode}
		l1Isolation = isolation.Numactl{PhyscpubindCPUs: defaultTopology.siblingThreadsToHpThreads.AvailableThreads().AsSlice(), PreferredNode: defaultTopology.numaNode}
		llcIsolation = isolation.Numactl{PhyscpubindCPUs: defaultTopology.sharingLLCButNotL1Threads.AsSlice(), PreferredNode: defaultTopology.numaNode}
		aggressorFactory = sensitivity.NewMultiIsolationAggressorFactory(l1Isolation, llcIsolation)
	}

	var hpExecutor executor.Executor
	var err error
	var memcachedLauncher memcached.Memcached
	memcachedConfig := memcached.DefaultMemcachedConfig()
	mutilateConfig := mutilate.DefaultMutilateConfig()

	if runOnKubernetesFlag.Value() {
		k8sConfig := kubernetes.DefaultConfig()
		k8sLauncher := kubernetes.New(executor.NewLocal(), executor.NewLocal(), k8sConfig)
		k8sClusterTaskHandle, err := k8sLauncher.Launch()
		errutil.Check(err)
		defer executor.StopCleanAndErase(k8sClusterTaskHandle)

		hpExecutorConfig := executor.DefaultKubernetesConfig()
		hpExecutorConfig.ContainerImage = "centos_swan_image"
		hpExecutorConfig.PodName = "swan-hp"
		hpExecutor, err = executor.NewKubernetes(hpExecutorConfig)
		errutil.Check(err)
	} else {
		hpExecutor = executor.NewLocalIsolated(hpIsolation)
	}

	memcachedLauncher = memcached.New(hpExecutor, memcachedConfig) // Initialize Memcached Launcher.

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

	// Initialize Mutilate Load Generator.
	mutilateLoadGenerator := mutilate.NewCluster(
		masterLoadGeneratorExecutor,
		agentsLoadGeneratorExecutors,
		mutilateConfig)

	// Initialize aggressors with BE isolation.
	aggressors := []sensitivity.LauncherSessionPair{}
	for _, aggressorName := range aggressorsFlag.Value() {
		aggressor, err := aggressorFactory.Create(aggressorName, runOnKubernetesFlag.Value())
		errutil.Check(err)

		aggressors = append(aggressors, sensitivity.NewLauncherWithoutSession(aggressor))
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
	err = sensitivityExperiment.Run()
	errutil.Check(err)
}
