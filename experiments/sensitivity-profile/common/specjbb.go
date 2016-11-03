package common

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/snap"

	"github.com/intelsdi-x/athena/pkg/snap/sessions/specjbb"
	"github.com/intelsdi-x/swan/experiments/sensitivity-profile/topology"
	"github.com/intelsdi-x/swan/experiments/sensitivity-profile/validate"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	specjbb_workload "github.com/intelsdi-x/swan/pkg/workloads/specjbb"
)

const (
	txICount = 1
)

// prepareSpecjbbLoadGenerator creates new LoadGenerator based on specjbb.
func prepareSpecjbbLoadGenerator() executor.LoadGenerator {
	specjbbLoadGeneratorConfig := specjbb_workload.NewDefaultConfig()
	specjbbLoadGeneratorConfig.TxICount = txICount

	var transactionInjectors []executor.Executor
	for i := 1; i <= txICount; i++ {
		transactionInjector := executor.NewLocal()
		transactionInjectors = append(transactionInjectors, transactionInjector)
	}
	loadGeneratorLauncher := specjbb_workload.NewLoadGenerator(executor.NewLocal(),
		transactionInjectors, specjbbLoadGeneratorConfig)

	return loadGeneratorLauncher
}

// repareSnapSpecjbbSessionLauncher prepare a SessionLauncher that runs SPECjbb collector and records that into storage.
// TODO: this should be put into athena:/pkg/snap
func prepareSnapSpecjbbSessionLauncher() (snap.SessionLauncher, error) {
	// NOTE: For debug it is convenient to disable snap for some experiment runs.
	if snap.SnapdHTTPEndpoint.Value() != "none" {
		// Create connection with Snap.
		logrus.Info("Connecting to Snapd on ", snap.SnapdHTTPEndpoint.Value())
		// TODO(bp): Make helper for passing host:port or only host option here.

		specjbbConfig := specjbbsession.DefaultConfig()
		specjbbConfig.SnapdAddress = snap.SnapdHTTPEndpoint.Value()
		specjbbSnapSession, err := specjbbsession.NewSessionLauncher(specjbbConfig)
		if err != nil {
			return nil, err
		}
		return specjbbSnapSession, nil
	}
	return nil, nil
}

// RunExperimentWithSpecjbbSessionLauncher prepares all the components necessary to run experiment.
// It uses specjbbSessionLauncherFactory to create a snap.SessionLauncher that will wrap specjbb (HP workload).
// Note: it includes parsing the environment to get configuration as well as preparing executors and eventually running the experiment.
func RunExperimentWithSpecjbbSessionLauncher(specjbbSessionLauncherFactory func(sensitivity.Configuration) snap.SessionLauncher) error {
	conf.SetAppName("specjbb-sensitivity-profile")
	conf.SetHelp(`Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
                     It executes workloads and triggers gathering of metrics like latency (SLI)`)
	logrus.SetLevel(conf.LogLevel())

	// Validate preconditions.
	validate.CheckCPUPowerGovernor()

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
	specjbbSessionLauncher := specjbbSessionLauncherFactory(configuration)

	// HP workload.
	backendConfig := specjbb_workload.DefaultSPECjbbBackendConfig()
	backendLauncher := specjbb_workload.NewBackend(hpExecutor, backendConfig)
	// NewMonitoredLauncher can accept nil as session launcher.
	backendLauncherSessionPair := sensitivity.NewMonitoredLauncher(backendLauncher, specjbbSessionLauncher)

	// Load generator.
	specjbbLoadGenerator := prepareSpecjbbLoadGenerator()

	specjbbSnapSession, err := prepareSnapSpecjbbSessionLauncher()
	if err != nil {
		return err
	}
	specjbbLoadGeneratorSessionPair := sensitivity.NewMonitoredLoadGenerator(specjbbLoadGenerator, specjbbSnapSession)

	// Experiment.
	sensitivityExperiment := sensitivity.NewExperiment(
		conf.AppName(),
		conf.LogLevel(),
		configuration,
		backendLauncherSessionPair,
		specjbbLoadGeneratorSessionPair,
		aggressorSessionLaunchers,
	)

	// Run experiment.
	return sensitivityExperiment.Run()
}
