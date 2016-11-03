package common

import (
	"runtime"

	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/isolation"
	"github.com/intelsdi-x/athena/pkg/kubernetes"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
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
	err := conf.ParseFlags()
	if err != nil {
		return err
	}
	switch hpTaskFlag.Value() {
	case "memcached":
		return RunExperimentWithMemcachedSessionLauncher(noopSessionLauncherFactory)
	case "specjbb":
		return RunExperimentWithSpecjbbSessionLauncher(noopSessionLauncherFactory)
	default:
		return RunExperimentWithMemcachedSessionLauncher(noopSessionLauncherFactory)
	}

}
