// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	RunOnKubernetesFlag = conf.NewBoolFlag("kubernetes", "Launch Kubernetes cluster and run workloads on Kubernetes. This flag is required to use other kubernetes flags. (caveat: cluster won't be started if `-kubernetes_run_on_existing` flag is set).  ", false)
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
		if !RunOnExistingKubernetesFlag.Value() {
			cleanup, err = LaunchKubernetesCluster()
			if err != nil {
				return nil, nil, nil, err
			}
		}

		hpExecutor, err = CreateKubernetesHpExecutor(hpIsolation)
		if err != nil {
			return nil, nil, nil, err
		}
		beExecutorFactory = DefaultKubernetesBEExecutorFactory
	} else {
		hpExecutor = executor.NewLocalIsolated(hpIsolation)
		beExecutorFactory = defaultLocalBEExecutorFactory
	}
	return
}

//LaunchKubernetesCluster starts new Kubernetes cluster using configuration provided with flags.
func LaunchKubernetesCluster() (cleanup func() error, err error) {
	k8sConfig := kubernetes.DefaultConfig()
	masterExecutor, err := executor.NewShell(k8sConfig.KubeAPIAddr)
	if err != nil {
		return nil, err
	}

	k8sLauncher := kubernetes.New(masterExecutor, executor.NewLocal(), k8sConfig)
	k8sClusterTaskHandle, err := k8sLauncher.Launch()
	if err != nil {
		return nil, err
	}

	cleanup = func() error {
		return k8sClusterTaskHandle.Stop()
	}

	return
}

//CreateKubernetesHpExecutor creates new instance of Kubernetes executor for HP task with isolation applied.
func CreateKubernetesHpExecutor(hpIsolation isolation.Decorator) (executor.Executor, error) {
	k8sConfig := kubernetes.DefaultConfig()
	k8sExecutorConfig := executor.DefaultKubernetesConfig()

	k8sExecutorConfig.ContainerImage = "centos_swan_image"
	k8sExecutorConfig.PodNamePrefix = "swan-hp"
	k8sExecutorConfig.NodeName = kubernetesNodeName.Value()
	k8sExecutorConfig.Decorators = isolation.Decorators{hpIsolation}
	k8sExecutorConfig.HostNetwork = true
	k8sExecutorConfig.Address = k8sConfig.GetKubeAPIAddress()
	k8sExecutorConfig.CPULimit = int64(HPKubernetesCPUResourceFlag.Value())
	k8sExecutorConfig.MemoryLimit = int64(HPKubernetesMemoryResourceFlag.Value())
	k8sExecutorConfig.CPURequest = k8sExecutorConfig.CPULimit
	k8sExecutorConfig.MemoryRequest = k8sExecutorConfig.MemoryLimit
	k8sExecutorConfig.Privileged = true

	return executor.NewKubernetes(k8sExecutorConfig)

}

//DefaultKubernetesBEExecutorFactory can be used to create Kubernetes executor for BE task with isolation applied.
func DefaultKubernetesBEExecutorFactory(decorators isolation.Decorators) (executor.Executor, error) {
	k8sConfig := kubernetes.DefaultConfig()
	config := executor.DefaultKubernetesConfig()
	config.PodNamePrefix = "swan-be"
	config.NodeName = kubernetesNodeName.Value()
	config.ContainerImage = "centos_swan_image"
	config.Decorators = decorators
	config.Privileged = true // swan aggressor use unshare, which requires sudo.
	config.Address = k8sConfig.GetKubeAPIAddress()
	return executor.NewKubernetes(config)
}

func defaultLocalBEExecutorFactory(decorators isolation.Decorators) (executor.Executor, error) {
	return executor.NewLocalIsolated(decorators), nil
}
