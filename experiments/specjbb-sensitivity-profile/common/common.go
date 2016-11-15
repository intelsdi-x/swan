package common

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/snap/sessions/specjbb"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	specjbb_workload "github.com/intelsdi-x/swan/pkg/workloads/specjbb"
)

var (
	specjbbIP = conf.NewIPFlag(
		"specjbb_loadgenerator_ip",
		"a",
		"127.0.0.1")
)

// prepareSpecjbbLoadGenerator creates new LoadGenerator based on specjbb.
func prepareSpecjbbLoadGenerator(ip string) (executor.LoadGenerator, error) {

	var loadGeneratorExecutor executor.Executor
	var transactionInjectors []executor.Executor
	var err error
	txICount := specjbb_workload.TxICountFlag.Value()
	if ip != "local" {
		loadGeneratorExecutor, err = sensitivity.NewRemote(ip)
		if err != nil {
			return nil, err
		}
		for i := 1; i <= txICount; i++ {
			transactionInjector, err := sensitivity.NewRemote(ip)
			if err != nil {
				return nil, err
			}
			transactionInjectors = append(transactionInjectors, transactionInjector)
		}
	} else {
		loadGeneratorExecutor = executor.NewLocal()
		for i := 1; i <= txICount; i++ {
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)
		}
	}

	specjbbLoadGeneratorConfig := specjbb_workload.NewDefaultConfig()
	specjbbLoadGeneratorConfig.ControllerIP = ip
	specjbbLoadGeneratorConfig.TxICount = txICount

	loadGeneratorLauncher := specjbb_workload.NewLoadGenerator(loadGeneratorExecutor,
		transactionInjectors, specjbbLoadGeneratorConfig)

	return loadGeneratorLauncher, nil
}

// repareSnapSpecjbbSessionLauncher prepare a SessionLauncher that runs SPECjbb collector and records that into storage.
// TODO: this should be put into athena:/pkg/snap
func prepareSnapSpecjbbSessionLauncher() (snap.SessionLauncher, error) {
	// NOTE: For debug it is convenient to disable snap for some experiment runs.
	if snap.SnapdHTTPEndpoint.Value() != "none" {
		// Create connection with Snap.
		logrus.Info("Connecting to Snapd on ", snap.SnapdHTTPEndpoint.Value())
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

// noopSessionLauncherFactory is a factory of snap.SessionLauncher that returns nothing.
func noopSessionLauncherFactory(_ sensitivity.Configuration) snap.SessionLauncher {
	return nil
}

// RunExperiment is main entrypoint to prepare and run experiment.
func RunExperiment() error {
	return RunExperimentWithSpecjbbSessionLauncher(noopSessionLauncherFactory)

}

// RunExperimentWithSpecjbbSessionLauncher prepares all the components necessary to run experiment.
// It uses specjbbSessionLauncherFactory to create a snap.SessionLauncher that will wrap specjbb (HP workload).
// Note: it includes parsing the environment to get configuration as well as preparing executors and eventually running the experiment.
func RunExperimentWithSpecjbbSessionLauncher(specjbbSessionLauncherFactory func(sensitivity.Configuration) snap.SessionLauncher) error {
	conf.SetAppName("specjbb-sensitivity-profile")
	conf.SetHelp(`Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
                     It executes workloads and triggers gathering of metrics like latency (SLI)`)
	err := conf.ParseFlags()
	if err != nil {
		return err
	}
	logrus.SetLevel(conf.LogLevel())

	specjbbHost := specjbbIP.Value()

	// Validate preconditions.
	validate.CheckCPUPowerGovernor()

	// Isolations.
	hpIsolation, l1Isolation, llcIsolation := topology.NewIsolations()

	// Executors.
	hpExecutor, beExecutorFactory, cleanup, err := sensitivity.PrepareExecutors(hpIsolation)
	if err != nil {
		return err
	}
	defer cleanup()

	// BE workloads.
	aggressorSessionLaunchers, err := sensitivity.PrepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	if err != nil {
		return err
	}

	// Prepare experiment configuration to be used by session launcher factory.
	configuration := sensitivity.DefaultConfiguration()
	specjbbSessionLauncher := specjbbSessionLauncherFactory(configuration)

	// HP workload.
	backendConfig := specjbb_workload.DefaultSPECjbbBackendConfig()
	backendConfig.IP = specjbbHost
	backendLauncher := specjbb_workload.NewBackend(hpExecutor, backendConfig)
	// NewMonitoredLauncher can accept nil as session launcher.
	backendLauncherSessionPair := sensitivity.NewMonitoredLauncher(backendLauncher, specjbbSessionLauncher)

	// Load generator.
	specjbbLoadGenerator, err := prepareSpecjbbLoadGenerator(specjbbHost)
	if err != nil {
		return err
	}
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
