package common

import (
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/isolation"
	"github.com/intelsdi-x/athena/pkg/kubernetes"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"

	"github.com/intelsdi-x/athena/pkg/snap/sessions/specjbb"
	"github.com/intelsdi-x/swan/experiments/specjbb-sensitivity-profile/topology"
	"github.com/intelsdi-x/swan/experiments/specjbb-sensitivity-profile/validate"
	specjbb_workload "github.com/intelsdi-x/swan/pkg/workloads/specjbb"
)

var (
	// Aggressors flag.
	aggressorsFlag = conf.NewSliceFlag(
		"aggr", "Aggressor to run experiment with. You can state as many as you want (--aggr=l1d --aggr=membw)")

	hpTaskFlag = conf.NewStringFlag("hp", "High priority task to run during experiment", "memcached")

	hpKubernetesCPUResourceFlag    = conf.NewIntFlag("hp_kubernetes_cpu_resource", "set limits & request for HP workloads pods run on kubernetes in CPU millis (default 1000 * number of CPU).", runtime.NumCPU()*1000)
	hpKubernetesMemoryResourceFlag = conf.NewIntFlag("hp_kubernetes_memory_resource", "set memory limits & request for HP pods workloads run on kubernetes in bytes (default 1GB).", 1000000000)

	runOnKubernetesFlag = conf.NewBoolFlag("run_on_kubernetes", "Launch HP and BE tasks on Kubernetes.", false)
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

// prepareAggressors prepare aggressors launcher's
// wrapped by session less pair using given isolations and executor factory for aggressor workloads.
// TODO: consider moving to swan:sensitivity/factory.go
func prepareAggressors(l1Isolation, llcIsolation isolation.Decorator, beExecutorFactory sensitivity.ExecutorFactoryFunc) (aggressorPairs []sensitivity.LauncherSessionPair, err error) {

	// Initialize aggressors with BE isolation wrapped as Snap session pairs.
	aggressorFactory := sensitivity.NewMultiIsolationAggressorFactory(l1Isolation, llcIsolation)

	for _, aggressorName := range aggressorsFlag.Value() {
		aggressorPair, err := aggressorFactory.Create(aggressorName, beExecutorFactory)
		if err != nil {
			return nil, err
		}
		aggressorPairs = append(aggressorPairs, sensitivity.NewLauncherWithoutSession(aggressorPair))
	}
	return
}

// prepareExecutors gives a executor to deploy your workloads with appliled isolation on HP.
func prepareExecutors(hpIsolation isolation.Decorator) (hpExecutor executor.Executor, beExecutorFactory sensitivity.ExecutorFactoryFunc, cleanup func(), err error) {
	if runOnKubernetesFlag.Value() {
		k8sConfig := kubernetes.DefaultConfig()
		k8sConfig.KubeAPIArgs = "--admission-control=\"AlwaysAdmit,AddToleration\"" // Enable AddToleration path by default.
		k8sLauncher := kubernetes.New(executor.NewLocal(), executor.NewLocal(), k8sConfig)
		k8sClusterTaskHandle, err := k8sLauncher.Launch()
		if err != nil {
			return nil, nil, nil, err
		}

		cleanup = func() { executor.StopCleanAndErase(k8sClusterTaskHandle) }

		// TODO: pass information from k8sConfig to hpExecutor and beExecutor configs.

		// HP executor.
		hpExecutorConfig := executor.DefaultKubernetesConfig()
		hpExecutorConfig.ContainerImage = "centos_swan_image"
		hpExecutorConfig.PodName = "swan-hp"
		hpExecutorConfig.Decorators = isolation.Decorators{hpIsolation}
		hpExecutorConfig.HostNetwork = true // requied to have access from mutilate agents run outside a k8s cluster.

		hpExecutorConfig.CPULimit = int64(hpKubernetesCPUResourceFlag.Value())
		hpExecutorConfig.MemoryLimit = int64(hpKubernetesMemoryResourceFlag.Value())
		// "Guranteed" class is when both resources and set for request and limit and equal.
		hpExecutorConfig.CPURequest = hpExecutorConfig.CPULimit
		hpExecutorConfig.MemoryRequest = hpExecutorConfig.MemoryLimit
		hpExecutor, err = executor.NewKubernetes(hpExecutorConfig)
		if err != nil {
			return nil, nil, nil, err
		}

		// BE Executors.
		beExecutorFactory = func(decorators isolation.Decorators) (executor.Executor, error) {
			config := executor.DefaultKubernetesConfig()
			config.ContainerImage = "centos_swan_image"
			config.Decorators = decorators
			config.PodName = "swan-aggr"
			config.Privileged = true // swan aggressor use unshare, which requires sudo.
			return executor.NewKubernetes(config)
		}
	} else {
		hpExecutor = executor.NewLocalIsolated(hpIsolation)
		cleanup = func() {}
		beExecutorFactory = func(decorators isolation.Decorators) (executor.Executor, error) {
			return executor.NewLocalIsolated(decorators), nil
		}
	}
	return
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
