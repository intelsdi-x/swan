package common

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/pkg/errors"
)

const (
	mutilateMasterFlagDefault = "127.0.0.1"
)

var (

	// Mutilate configuration.
	percentileFlag     = conf.NewStringFlag("percentile", "Tail latency Percentile", "99")
	mutilateMasterFlag = conf.NewStringFlag(
		"mutilate_master",
		"Mutilate master host for remote executor. In case of 0 agents being specified it runs in agentless mode."+
			"Use `local` to run with local executor.",
		"127.0.0.1")
	mutilateAgentsFlag = conf.NewSliceFlag(
		"mutilate_agent",
		"Mutilate agent hosts for remote executor. Can be specified many times for multiple agents setup.")
)

// PrepareSnapMutilateSessionLauncher prepares a SessionLauncher that runs mutilate collector and records that into storage.
// TODO: this should be put into swan:/pkg/snap
func PrepareSnapMutilateSessionLauncher() (snap.SessionLauncher, error) {
	// Create connection with Snap.
	logrus.Info("Connecting to Snapteld on ", snap.SnapteldHTTPEndpoint.Value())
	// TODO(bp): Make helper for passing host:port or only host option here.

	mutilateConfig := mutilatesession.DefaultConfig()
	mutilateConfig.SnapteldAddress = snap.SnapteldHTTPEndpoint.Value()
	mutilateSnapSession, err := mutilatesession.NewSessionLauncher(mutilateConfig)
	if err != nil {
		return nil, err
	}
	return mutilateSnapSession, nil
}

// PrepareMutilateGenerator creates new LoadGenerator based on mutilate.
func PrepareMutilateGenerator(memcacheIP string, memcachePort int) (executor.LoadGenerator, error) {
	mutilateConfig := mutilate.DefaultMutilateConfig()
	mutilateConfig.MemcachedHost = memcacheIP
	mutilateConfig.MemcachedPort = memcachePort
	mutilateConfig.LatencyPercentile = percentileFlag.Value()

	// Special case to have ability to use local executor for mutilate master load generator.
	// This is needed for docker testing.
	var masterLoadGeneratorExecutor executor.Executor
	agentsLoadGeneratorExecutors := []executor.Executor{}
	if mutilateMasterFlag.Value() != mutilateMasterFlagDefault {
		var err error
		masterLoadGeneratorExecutor, err = executor.NewRemoteFromIP(mutilateMasterFlag.Value())
		if err != nil {
			return nil, err
		}
	} else {
		masterLoadGeneratorExecutor = executor.NewLocal()
	}
	// Pack agents.
	for _, agent := range mutilateAgentsFlag.Value() {
		remoteExecutor, err := executor.NewRemoteFromIP(agent)
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

// GetPeakLoad runs tuning in order to determine the peak load.
func GetPeakLoad(hpLauncher executor.Launcher, loadGenerator executor.LoadGenerator, slo int) (peakLoad int, err error) {
	prTask, err := hpLauncher.Launch()
	if err != nil {
		return 0, errors.Wrap(err, "cannot launch memcached")
	}
	defer func() {
		// If function terminated with error then we do not want to overwrite it with any errors in defer.
		errStop := prTask.Stop()
		if err == nil {
			err = errStop
		}
		prTask.Clean()
	}()

	err = loadGenerator.Populate()
	if err != nil {
		return 0, errors.Wrap(err, "cannot populate memcached")
	}

	peakLoad, _, err = loadGenerator.Tune(slo)
	if err != nil {
		return 0, errors.Wrap(err, "tuning failed")
	}

	return
}
