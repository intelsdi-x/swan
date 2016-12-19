package common

import (
	"os"
	"path"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
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
// TODO: this should be put into athena:/pkg/snap
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
		masterLoadGeneratorExecutor, err = sensitivity.NewRemote(mutilateMasterFlag.Value())
		if err != nil {
			return nil, err
		}
		// Pack agents.
		for _, agent := range mutilateAgentsFlag.Value() {
			remoteExecutor, err := sensitivity.NewRemote(agent)
			if err != nil {
				return nil, err
			}
			agentsLoadGeneratorExecutors = append(agentsLoadGeneratorExecutors, remoteExecutor)
		}
	} else {
		masterLoadGeneratorExecutor = executor.NewLocal()
		for range mutilateAgentsFlag.Value() {
			remoteExecutor := executor.NewLocal()
			agentsLoadGeneratorExecutors = append(agentsLoadGeneratorExecutors, remoteExecutor)
		}
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

// CreateRepetitionDir creates folders that store repetition logs.
func CreateRepetitionDir(experimentDirectory, phaseName string, repetition int) (string, error) {
	repetitionDir := path.Join(experimentDirectory, phaseName, strconv.Itoa(repetition))
	err := os.MkdirAll(repetitionDir, 0777)
	if err != nil {
		return "", errors.Wrapf(err, "could not create dir %q", repetitionDir)
	}

	err = os.Chdir(repetitionDir)
	if err != nil {
		return "", errors.Wrapf(err, "could not change to dir %q", repetitionDir)
	}

	return repetitionDir, nil
}

// CreateExperimentDir creates directory structure for the experiment.
func CreateExperimentDir(uuid string) (experimentDirectory string, logFile *os.File, err error) {
	experimentDirectory = path.Join(os.TempDir(), conf.AppName(), uuid)
	err = os.MkdirAll(experimentDirectory, 0777)
	if err != nil {
		return "", &os.File{}, errors.Wrapf(err, "cannot create experiment directory: ", experimentDirectory)
	}
	err = os.Chdir(experimentDirectory)
	if err != nil {
		return "", &os.File{}, errors.Wrapf(err, "cannot chdir to experiment directory", experimentDirectory)
	}

	masterLogFilename := path.Join(experimentDirectory, "master.log")
	logFile, err = os.OpenFile(masterLogFilename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return "", &os.File{}, errors.Wrapf(err, "could not open log file %q", masterLogFilename)
	}

	return experimentDirectory, logFile, nil
}

// GetPeakLoad runs tuning in order to determine the peak load.
func GetPeakLoad(hpLauncher executor.Launcher, loadGenerator executor.LoadGenerator, slo int) (int, error) {
	prTask, err := hpLauncher.Launch()
	if err != nil {
		return 0, errors.Wrap(err, "cannot launch memcached")
	}
	defer func() {
		prTask.Stop()
		prTask.Clean()
	}()

	err = loadGenerator.Populate()
	if err != nil {
		return 0, errors.Wrap(err, "cannot populate memcached")
	}

	load, _, err := loadGenerator.Tune(slo)
	if err != nil {
		return 0, errors.Wrap(err, "tuning failed")
	}

	return int(load), nil
}
