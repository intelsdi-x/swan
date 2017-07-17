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
	"github.com/intelsdi-x/swan/pkg/experiment"
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

	// hpKubernetesCPUResourceFlag indicates CPU shares that HP task should be allowed to use.
	hpKubernetesCPUResourceFlag = conf.NewIntFlag("kubernetes_hp_cpu_resource", "Sets CPU resource limit and request for HP workload on Kubernetes [CPU millis, default 1000 * number of CPU].", runtime.NumCPU()*1000)
	// hpKubernetesMemoryResourceFlag indicates amount of memory that HP task can use.
	hpKubernetesMemoryResourceFlag = conf.NewIntFlag("kubernetes_hp_memory_resource", "Sets memory limit and request for HP workloads on Kubernetes in bytes (default 4GB).", 4000000000)

	// hpKubernetesGuaranteedClassFlag indicates tha HP workload will run as guarateed class.
	hpKubernetesGuaranteedClassFlag = conf.NewBoolFlag("kubernetes_hp_guaranteed_class", "Run HP workload on Kubernetes as Pod with \"QoS Guaranteed resources class\" (by default runs as \"Burstable class\").", false)

	kubernetesNodeName = conf.NewStringFlag("kubernetes_target_node_name", fmt.Sprintf("Experiment's Kubernetes pods will be run on this node. Helpful when used with %q flag. Default is `$HOSTNAME`", experiment.RunOnExistingKubernetesFlag.Name), hostname)
)

// ExecutorFactory is prepares executor for High Priority and Best Effort workloads.
type ExecutorFactory interface {
	// BuildHighPriorityExecutor returns executor for High Priority workloads.
	BuildHighPriorityExecutor(decorator ...isolation.Decorator) (executor.Executor, error)
	// BuildBestEffortExecutor returns executor for Best Effort workloads.
	BuildBestEffortExecutor(decorator ...isolation.Decorator) (executor.Executor, error)
}

// NewExecutorFactory returns Local or Kubernetes executor factory, depending on flags.
func NewExecutorFactory() ExecutorFactory {
	if experiment.RunOnKubernetesFlag.Value() {
		return NewKubernetesExecutorFactory()
	}

	return NewLocalExecutorFactory()
}

// LocalExecutorFactory produces local executors.
type LocalExecutorFactory struct {
}

// NewLocalExecutorFactory returns Local Executor Factory instance.
func NewLocalExecutorFactory() ExecutorFactory {
	return &LocalExecutorFactory{}
}

// BuildHighPriorityExecutor returns local executor.
func (factory LocalExecutorFactory) BuildHighPriorityExecutor(decorators ...isolation.Decorator) (executor.Executor, error) {
	return executor.NewLocalIsolated(decorators...), nil
}

// BuildBestEffortExecutor returns local executor.
func (factory LocalExecutorFactory) BuildBestEffortExecutor(decorators ...isolation.Decorator) (executor.Executor, error) {
	return executor.NewLocalIsolated(decorators...), nil
}

// KubernetesExecutorFactory produces Kubernetes Executors.
type KubernetesExecutorFactory struct {
}

// NewKubernetesExecutorFactory returns Kubernetes Executor Factory instance.
func NewKubernetesExecutorFactory() ExecutorFactory {
	return &KubernetesExecutorFactory{}
}

// BuildHighPriorityExecutor returns Kubernetes Executor with Guaranteed or Burstable QoS class (depending on hpKubernetesGuaranteedClassFlag).
func (factory KubernetesExecutorFactory) BuildHighPriorityExecutor(decorators ...isolation.Decorator) (executor.Executor, error) {
	clusterConfig := kubernetes.DefaultConfig()
	k8sExecutorConfig := executor.DefaultKubernetesConfig()

	k8sExecutorConfig.PodNamePrefix = "swan-hp"
	k8sExecutorConfig.NodeName = kubernetesNodeName.Value()
	k8sExecutorConfig.Decorators = decorators
	k8sExecutorConfig.HostNetwork = true
	k8sExecutorConfig.Address = clusterConfig.GetKubeAPIAddress()
	k8sExecutorConfig.CPURequest = int64(hpKubernetesCPUResourceFlag.Value())
	k8sExecutorConfig.MemoryRequest = int64(hpKubernetesMemoryResourceFlag.Value())

	// Create guaranteed (resource requests == limits) pod.
	if hpKubernetesGuaranteedClassFlag.Value() {
		k8sExecutorConfig.CPULimit = int64(hpKubernetesCPUResourceFlag.Value())
		k8sExecutorConfig.MemoryLimit = int64(hpKubernetesMemoryResourceFlag.Value())
	}

	k8sExecutorConfig.Privileged = true

	return executor.NewKubernetes(k8sExecutorConfig)
}

// BuildBestEffortExecutor returns executor with Best Effort QoS class.
func (factory KubernetesExecutorFactory) BuildBestEffortExecutor(decorators ...isolation.Decorator) (executor.Executor, error) {
	clusterConfig := kubernetes.DefaultConfig()

	config := executor.DefaultKubernetesConfig()
	config.Address = clusterConfig.GetKubeAPIAddress()
	config.PodNamePrefix = "swan-be"
	config.NodeName = kubernetesNodeName.Value()
	config.Decorators = decorators
	config.Privileged = true // Best Effort workloads use unshare, which requires sudo.
	return executor.NewKubernetes(config)
}
