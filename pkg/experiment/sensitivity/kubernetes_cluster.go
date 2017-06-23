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
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
)

var (
	// runOnKubernetesFlag indicates that experiment is to be run on K8s cluster.
	runOnKubernetesFlag = conf.NewBoolFlag("kubernetes", "Launch Kubernetes cluster and run workloads on Kubernetes. This flag is required to use other kubernetes flags. (caveat: cluster won't be started if `-kubernetes_run_on_existing` flag is set).  ", false)
	// runOnExistingKubernetesFlag indicates that experiment should not set up a Kubernetes cluster but use an existing one.
	runOnExistingKubernetesFlag = conf.NewBoolFlag("kubernetes_run_on_existing", "Launch HP and BE tasks on existing Kubernetes cluster. (It has to be used with --kubernetes flag). User should provide 'kubernetes_kubeconfig' flag to kubeconfig to point proper API server.", false)
)

// ShouldLaunchKubernetesCluster checks runOnKubernetesFlag and runOnExistingKubernetesFlag
// and returns information if Kubernetes cluster should be launched.
func ShouldLaunchKubernetesCluster() bool {
	return runOnKubernetesFlag.Value() == true && runOnExistingKubernetesFlag.Value() == false
}

//LaunchKubernetesCluster starts new Kubernetes cluster using configuration provided with flags.
func LaunchKubernetesCluster() (clusterHandle executor.TaskHandle, err error) {
	masterExecutor, err := executor.NewShell(kubernetes.KubernetesMasterFlag.Value())
	if err != nil {
		return nil, err
	}

	k8sLauncher := kubernetes.New(masterExecutor, executor.NewLocal(), kubernetes.DefaultConfig())
	k8sClusterTaskHandle, err := k8sLauncher.Launch()
	if err != nil {
		return nil, err
	}

	return k8sClusterTaskHandle, nil
}
