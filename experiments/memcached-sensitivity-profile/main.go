package main

import (
	"os/user"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/kubesnap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
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

	kubernetesMetricsFlag = conf.NewSliceFlag("kubernetes_metrics", "Select kubesnap metrics")

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
	var hpIsolation, beIsolation, l1Isolation, llcIsolation isolation.Isolation
	var aggressorFactory sensitivity.AggressorFactory

	if runOnKubernetesFlag.Value() {
		var dummyIsolation DecoratorFunc = func(s string) string { return s }
		aggressorFactory = sensitivity.NewSingleIsolationAggressorFactory(dummyIsolation)
	} else if isManualPolicy() {
		hpIsolation, beIsolation = manualPolicy()
		aggressorFactory = sensitivity.NewSingleIsolationAggressorFactory(beIsolation)
		defer beIsolation.Clean()
		defer hpIsolation.Clean()
	} else {
		// NOTE: Temporary hack for having multiple isolations in Sensitivity Profile.
		hpIsolation, l1Isolation, llcIsolation = sensitivityProfileIsolationPolicy()
		aggressorFactory = sensitivity.NewMultiIsolationAggressorFactory(l1Isolation, llcIsolation)
		defer l1Isolation.Clean()
		defer llcIsolation.Clean()
		defer hpIsolation.Clean()
	}

	var hpExecutor executor.Executor
	var err error
	var kubesnapLauncher *kubesnap.SessionLauncher
	var memcachedLauncher memcached.Memcached

	memcachedConfig := memcached.DefaultMemcachedConfig()
	mutilateConfig := mutilate.DefaultMutilateConfig()

	if runOnKubernetesFlag.Value() {
		k8sConfig, err := kubernetes.DefaultConfig()
		errutil.Check(err)
		k8sLauncher := kubernetes.New(executor.NewLocal(), executor.NewLocal(), k8sConfig)
		k8sClusterTaskHandle, err := k8sLauncher.Launch()
		defer executor.StopCleanAndErase(k8sClusterTaskHandle)

		hpExecutorConfig := executor.DefaultKubernetesConfig()
		hpExecutorConfig.ContainerImage = "centos_swan_image"
		hpExecutorConfig.PodName = "swan-hp"
		hpExecutor, err = executor.NewKubernetes(hpExecutorConfig)

		kubesnapConfig := kubesnap.DefaultConfig()
		if len(kubernetesMetricsFlag.Value()) > 0 {
			kubesnapConfig.Metrics = kubernetesMetricsFlag.Value()
		}
		kubesnapLauncher, err = kubesnap.NewSessionLauncher(kubesnapConfig)

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



	if runOnKubernetesFlag.Value() {
		expSession := phase.Session{
			PhaseID: "launch",
			AggressorName: "",
			ExperimentID: "experiment",
			LoadPointQPS: 0,
			RepetitionID: 0,
		}
		kubesnapLauncher.LaunchSession(nil, expSession)
	}

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
