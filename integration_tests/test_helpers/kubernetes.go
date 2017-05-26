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

package testhelpers

import (
	"encoding/json"
	"fmt"

	"github.com/intelsdi-x/swan/pkg/executor"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

// KubeClient is a helper struct to communicate with K8s API. It stores
// Kubernetes client and extends it to additional functionality needed
// by integration tests.
type KubeClient struct {
	Clientset *kubernetes.Clientset
	namespace string
}

// NewKubeClient creates KubeClient object based on given KubernetesConfig
// structure. It returns error if given configuration is invalid.
func NewKubeClient(kubernetesConfig executor.KubernetesConfig) (*KubeClient, error) {
	kubectlConfig := &rest.Config{
		Host: kubernetesConfig.Address,
	}

	cli, err := kubernetes.NewForConfig(kubectlConfig)
	if err != nil {
		return nil, err
	}
	return &KubeClient{
		Clientset: cli,
		namespace: kubernetesConfig.Namespace,
	}, nil
}

// GetPods gathers running and not running pods from K8s cluster.
func (k *KubeClient) GetPods() ([]*v1.Pod, []*v1.Pod, error) {
	pods, err := k.Clientset.Pods(k.namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	var runningPods []*v1.Pod
	var notRunningPods []*v1.Pod

	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case v1.PodRunning:
			runningPods = append(runningPods, &pod)
		case v1.PodPending:
		case v1.PodSucceeded:
		case v1.PodFailed:
			notRunningPods = append(notRunningPods, &pod)
		case v1.PodUnknown:
			return nil, nil, fmt.Errorf("at least one of pods is in Unknown state")
		}
	}

	return runningPods, notRunningPods, nil
}

// DeletePod with given podName.
func (k *KubeClient) DeletePod(podName string) error {
	var oneSecond int64 = 1
	return k.Clientset.Pods(k.namespace).Delete(podName, &v1.DeleteOptions{GracePeriodSeconds: &oneSecond})
}

// Node assume just one node a return it. Note panics if unavailable (this is just test helper!).
func (k *KubeClient) node() *v1.Node {
	nodes, err := k.Clientset.Nodes().List(v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	if len(nodes.Items) != 1 {
		panic("Expected signle nodes kubernetes cluster!")
	}
	return &nodes.Items[0]
}

// TaintNode with NoSchedule taint (can panic).
func (k *KubeClient) TaintNode() {
	newTaint := api.Taint{
		Key: "hponly", Value: "true", Effect: api.TaintEffectNoSchedule,
	}
	taintsInJSON, err := json.Marshal([]api.Taint{newTaint})
	if err != nil {
		panic(err)
	}

	node := k.node()
	k.updateTaints(node, taintsInJSON)
}

func (k *KubeClient) updateTaints(node *v1.Node, taints []byte) {
	patchSet := v1.Node{}
	patchSet.Annotations = map[string]string{api.TaintsAnnotationKey: string(taints)}
	patchSetInJSON, err := json.Marshal(patchSet)
	if err != nil {
		panic(err)
	}
	_, err = k.Clientset.Nodes().Patch(node.Name, api.MergePatchType, patchSetInJSON)
	if err != nil {
		panic(err)
	}
}

// UntaintNode removes all tains for given node (can panic on failure).
func (k *KubeClient) UntaintNode() {
	taintsInJSON, err := json.Marshal([]api.Taint{})
	if err != nil {
		panic(err)
	}
	node := k.node()
	k.updateTaints(node, taintsInJSON)
}
