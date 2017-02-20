package sensitivity

import (
	"fmt"
	"runtime"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
)

var (
	hpKubernetesCPUResourceFlag    = conf.NewIntFlag("hp_kubernetes_cpu_resource", "set limits and request for HP workloads pods run on kubernetes in CPU millis (default 1000 * number of CPU).", runtime.NumCPU()*1000)
	hpKubernetesMemoryResourceFlag = conf.NewIntFlag("hp_kubernetes_memory_resource", "set memory limits and request for HP pods workloads run on kubernetes in bytes (default 1GB).", 1000000000)

	runOnKubernetesFlag = conf.NewBoolFlag("kubernetes", "Launch HP and BE tasks on Kubernetes.", false)
	runOnExistingKubernetesFlag = conf.NewBoolFlag("kubernetes_run_on_existing", "Launch HP and BE tasks on existing Kubernetes cluster. (can be use only with --kubernetes flag)", false)

	// KubernetesMasterFlag represents address of a host where Kubernetes master components are to be run
	KubernetesMasterFlag = conf.NewStringFlag("kubernetes_master", "Address of a host where Kubernetes master components are to be run", "127.0.0.1")
)

// PrepareExecutors gives an executor to deploy your workloads with applied isolation on HP.
func PrepareExecutors(hpIsolation isolation.Decorator) (hpExecutor executor.Executor, beExecutorFactory ExecutorFactoryFunc, cleanup func(), err error) {
	cleanup = func() {}
	if runOnKubernetesFlag.Value() {
		k8sConfig := kubernetes.DefaultConfig()
		masterAddress := KubernetesMasterFlag.Value()
		apiAddress := fmt.Sprintf("%s:%s", masterAddress, "8080")
		masterExecutor, err := executor.NewRemoteFromIP(masterAddress)
		if err != nil {
			return nil, nil, nil, err
		}

		if !runOnExistingKubernetesFlag.Value() {
			k8sLauncher := kubernetes.New(masterExecutor, executor.NewLocal(), k8sConfig)
			k8sClusterTaskHandle, err := k8sLauncher.Launch()
			if err != nil {
				return nil, nil, nil, err
			}

			cleanup = func() {
				executor.StopCleanAndErase(k8sClusterTaskHandle)
			}
		}

		// TODO: pass information from k8sConfig to hpExecutor and beExecutor configs.

		// HP executor.
		hpExecutorConfig := executor.DefaultKubernetesConfig()
		hpExecutorConfig.ContainerImage = "centos_swan_image"
		hpExecutorConfig.PodNamePrefix = "swan-hp"
		hpExecutorConfig.Decorators = isolation.Decorators{hpIsolation}
		hpExecutorConfig.HostNetwork = true // requied to have access from mutilate agents run outside a k8s cluster.
		hpExecutorConfig.Address = apiAddress

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
			config.PodNamePrefix = "swan-be"
			config.ContainerImage = "centos_swan_image"
			config.Decorators = decorators
			config.Privileged = true // swan aggressor use unshare, which requires sudo.
			config.Address = apiAddress
			return executor.NewKubernetes(config)
		}
	} else {
		hpExecutor = executor.NewLocalIsolated(hpIsolation)
		beExecutorFactory = func(decorators isolation.Decorators) (executor.Executor, error) {
			return executor.NewLocalIsolated(decorators), nil
		}
	}
	return
}
