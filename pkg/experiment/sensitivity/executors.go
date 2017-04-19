package sensitivity

import (
	"fmt"
	"os"
	"runtime"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
)

var (
	hostname = func() string {
		hostname, err := os.Hostname()
		if err != nil {
			panic(fmt.Sprintf("%s", err.Error()))
		}
		return hostname
	}()

	// RunOnKubernetesFlag indicates that experiment is to be run on K8s cluster.
	RunOnKubernetesFlag = conf.NewBoolFlag("kubernetes", "Launch Kubernets cluster and workload on Kubernetes. This flag is required to use other kubernetes flags.", false)
	// RunOnExistingKubernetesFlag indicates that experiment should not set up a Kubernetes cluster but use an existing one.
	RunOnExistingKubernetesFlag = conf.NewBoolFlag("kubernetes_run_on_existing", "Launch HP and BE tasks on existing Kubernetes cluster. (It has to be used with --kubernetes flag). User should provide 'kubernetes_kubeconfig' flag to kubeconfig to point proper API server.", false)

	// HPKubernetesCPUResourceFlag indicates CPU shares that HP task should be allowed to use.
	HPKubernetesCPUResourceFlag = conf.NewIntFlag("kubernetes_hp_cpu_resource", "Sets CPU resource limit and request for HP workload on Kubernetes [CPU millis, default 1000 * number of CPU].", runtime.NumCPU()*1000)
	// HPKubernetesMemoryResourceFlag indicates amount of memory that HP task can use.
	HPKubernetesMemoryResourceFlag = conf.NewIntFlag("kubernetes_hp_memory_resource", "Sets memory limit and request for HP workloads on Kubernetes in bytes (default 1GB).", 1000000000)

	kubernetesNodeName = conf.NewStringFlag("kubernetes_target_node_name", "Experiment's Kubernetes pods will be run on this node.", hostname)
)

// PrepareExecutors gives an executor to deploy your workloads with applied isolation on HP.
func PrepareExecutors(hpIsolation isolation.Decorator) (hpExecutor executor.Executor, beExecutorFactory ExecutorFactoryFunc, cleanup func() error, err error) {
	if RunOnKubernetesFlag.Value() {
		k8sConfig := kubernetes.DefaultConfig()

		if !RunOnExistingKubernetesFlag.Value() {
			masterExecutor, err := executor.NewRemoteFromIP(k8sConfig.KubeAPIAddr)
			if err != nil {
				return nil, nil, nil, err
			}
			k8sLauncher := kubernetes.New(masterExecutor, executor.NewLocal(), k8sConfig)
			k8sClusterTaskHandle, err := k8sLauncher.Launch()
			if err != nil {
				return nil, nil, nil, err
			}

			cleanup = func() error {
				err := executor.StopAndEraseOutput(k8sClusterTaskHandle)
				return err.GetErrIfAny()
			}
		}

		// TODO: pass information from k8sConfig to hpExecutor and beExecutor configs.

		// HP executor.
		hpExecutorConfig := executor.DefaultKubernetesConfig()
		hpExecutorConfig.NodeName = kubernetesNodeName.Value()
		hpExecutorConfig.ContainerImage = "centos_swan_image"
		hpExecutorConfig.PodNamePrefix = "swan-hp"
		hpExecutorConfig.Decorators = isolation.Decorators{hpIsolation}
		hpExecutorConfig.HostNetwork = true // requied to have access from mutilate agents run outside a k8s cluster.
		hpExecutorConfig.Address = k8sConfig.GetKubeAPIAddress()
		hpExecutorConfig.CPULimit = int64(HPKubernetesCPUResourceFlag.Value())
		hpExecutorConfig.MemoryLimit = int64(HPKubernetesMemoryResourceFlag.Value())
		// "Guranteed" class is when both resources and set for request and limit and equal.
		hpExecutorConfig.CPURequest = hpExecutorConfig.CPULimit
		hpExecutorConfig.MemoryRequest = hpExecutorConfig.MemoryLimit
		hpExecutorConfig.Privileged = true
		hpExecutor, err = executor.NewKubernetes(hpExecutorConfig)
		if err != nil {
			return nil, nil, nil, err
		}

		// BE Executors.
		beExecutorFactory = func(decorators isolation.Decorators) (executor.Executor, error) {
			config := executor.DefaultKubernetesConfig()
			config.PodNamePrefix = "swan-be"
			config.ContainerImage = "centos_swan_image"
			config.NodeName = kubernetesNodeName.Value()
			config.Decorators = decorators
			config.Privileged = true // swan aggressor use unshare, which requires sudo.
			config.Address = k8sConfig.GetKubeAPIAddress()
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
