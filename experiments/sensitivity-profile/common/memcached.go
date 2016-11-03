package common

import (
	"os/user"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/experiments/sensitivity-profile/topology"
	"github.com/intelsdi-x/swan/experiments/sensitivity-profile/validate"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	memcached_workload "github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
)

const (
	mutilateMasterFlagDefault = "local"
)

var (

	// Mutilate configuration.
	percentileFlag     = conf.NewStringFlag("percentile", "Tail latency Percentile", "99")
	mutilateMasterFlag = conf.NewIPFlag(
		"mutilate_master",
		"Mutilate master host for remote executor. In case of 0 agents being specified it runs in agentless mode."+
			"Use `local` to run with local executor.",
		"127.0.0.1")
	mutilateAgentsFlag = conf.NewSliceFlag(
		"mutilate_agent",
		"Mutilate agent hosts for remote executor. Can be specified many times for multiple agents setup.")
)

// prepareMutilateGenerator create new LoadGenerator based on mutilate.
func prepareMutilateGenerator(memcacheIP string, memcachePort int) (executor.LoadGenerator, error) {
	mutilateConfig := mutilate.DefaultMutilateConfig()
	mutilateConfig.MemcachedHost = memcacheIP
	mutilateConfig.MemcachedPort = memcachePort
	mutilateConfig.LatencyPercentile = percentileFlag.Value()

	// Special case to have ability to use local executor for mutilate master load generator.
	// This is needed for docker testing.
	var masterLoadGeneratorExecutor executor.Executor
	masterLoadGeneratorExecutor = executor.NewLocal()
	if mutilateMasterFlag.Value() != mutilateMasterFlagDefault {
		var err error
		masterLoadGeneratorExecutor, err = newRemote(mutilateMasterFlag.Value())
		if err != nil {
			return nil, err
		}
	}

	// Pack agents.
	agentsLoadGeneratorExecutors := []executor.Executor{}
	for _, agent := range mutilateAgentsFlag.Value() {
		remoteExecutor, err := newRemote(agent)
		if err != nil {
			return nil, err
		}
		agentsLoadGeneratorExecutors = append(agentsLoadGeneratorExecutors, remoteExecutor)
	}
	logrus.Debugf("Added %d mutilate agent(s) to mutilate cluster", len(agentsLoadGeneratorExecutors))

	// Validate mutilate cluster executors and their limit of
	// number of open file descriptors. Sane mutilate configuration requires
	// more than default (1024) for mutilate cluster.
	validate.ExecutorsNOFILELimit(
		append(agentsLoadGeneratorExecutors, masterLoadGeneratorExecutor),
	)

	// Initialize Mutilate Load Generator.
	mutilateLoadGenerator := mutilate.NewCluster(
		masterLoadGeneratorExecutor,
		agentsLoadGeneratorExecutors,
		mutilateConfig)

	return mutilateLoadGenerator, nil
}

// newRemote is helper for creating remotes with default sshConfig.
// TODO: this should be put into athena:/pkg/executors
func newRemote(ip string) (executor.Executor, error) {
	// TODO(bp): Have ability to choose user using parameter here.
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	sshConfig, err := executor.NewSSHConfig(ip, executor.DefaultSSHPort, user)
	if err != nil {
		return nil, err
	}

	return executor.NewRemote(sshConfig), nil
}

// PrepareSnapMutilateSessionLauncher prepare a SessionLauncher that runs mutilate collector and records that into storage.
// Note: SnapdHTTPEndpoint set to "none" will disable mutilate session completely.
// TODO: this should be put into athena:/pkg/snap
func prepareSnapMutilateSessionLauncher() (snap.SessionLauncher, error) {
	// NOTE: For debug it is convenient to disable snap for some experiment runs.
	if snap.SnapdHTTPEndpoint.Value() != "none" {
		// Create connection with Snap.
		logrus.Info("Connecting to Snapd on ", snap.SnapdHTTPEndpoint.Value())
		// TODO(bp): Make helper for passing host:port or only host option here.

		mutilateConfig := mutilatesession.DefaultConfig()
		mutilateConfig.SnapdAddress = snap.SnapdHTTPEndpoint.Value()
		mutilateSnapSession, err := mutilatesession.NewSessionLauncher(mutilateConfig)
		if err != nil {
			return nil, err
		}
		return mutilateSnapSession, nil
	}
	return nil, nil
}

// RunExperimentWithMemcachedSessionLauncher is preparing all the components necessary to run experiment but uses memcachedSessionLauncherFactory
// to create a snap.SessionLauncher that will wrap memcached (HP workload).
// Note: it includes parsing the environment to get configuration as well as preparing executors and eventually running the experiment.
func RunExperimentWithMemcachedSessionLauncher(memcachedSessionLauncherFactory func(sensitivity.Configuration) snap.SessionLauncher) error {
	conf.SetAppName("memcached-sensitivity-profile")
	conf.SetHelp(`Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
It executes workloads and triggers gathering of certain metrics like latency (SLI) and the achieved number of Request per Second (QPS/RPS)`)
	logrus.SetLevel(conf.LogLevel())

	// Validate preconditions.
	validate.OS()

	// Isolations.
	hpIsolation, l1Isolation, llcIsolation := topology.NewIsolations()

	// Executors.
	hpExecutor, beExecutorFactory, cleanup, err := prepareExecutors(hpIsolation)
	if err != nil {
		return err
	}
	defer cleanup()

	// BE workloads.
	aggressorSessionLaunchers, err := prepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	if err != nil {
		return err
	}

	// Prepare experiment configuration to be used by session launcher factory.
	configuration := sensitivity.DefaultConfiguration()
	memcachedSessionLauncher := memcachedSessionLauncherFactory(configuration)

	// HP workload.
	memcachedConfig := memcached_workload.DefaultMemcachedConfig()
	memcachedLauncher := memcached_workload.New(hpExecutor, memcachedConfig)
	memcachedLauncherSessionPair := sensitivity.NewMonitoredLauncher(memcachedLauncher, memcachedSessionLauncher) // NewMonitoredLauncher can accept nil as session launcher.

	// Load generator.
	mutilateLoadGenerator, err := prepareMutilateGenerator(memcachedConfig.IP, memcachedConfig.Port)
	if err != nil {
		return err
	}

	mutilateSnapSession, err := prepareSnapMutilateSessionLauncher()
	if err != nil {
		return err
	}
	mutilateLoadGeneratorSessionPair := sensitivity.NewMonitoredLoadGenerator(mutilateLoadGenerator, mutilateSnapSession)

	// Experiment.
	sensitivityExperiment := sensitivity.NewExperiment(
		conf.AppName(),
		conf.LogLevel(),
		configuration,
		memcachedLauncherSessionPair,
		mutilateLoadGeneratorSessionPair,
		aggressorSessionLaunchers,
	)

	// Run experiment.
	return sensitivityExperiment.Run()
}
