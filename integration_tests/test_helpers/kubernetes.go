package testhelpers

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/athena/pkg/executor"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

// KubeClient is a helper struct to communicate with K8s API. It stores
// Kubernetes client and extends it to additional functionality needed
// by integration tests.
type KubeClient struct {
	*client.Client
	namespace string
}

// NewKubeClient creates KubeClient object based on given KubernetesConfig
// structure. It returns error if given configuration is invalid.
func NewKubeClient(kubernetesConfig executor.KubernetesConfig) (*KubeClient, error) {
	kubectlConfig := &restclient.Config{
		Host:     kubernetesConfig.Address,
		Username: kubernetesConfig.Username,
		Password: kubernetesConfig.Password,
	}

	cli, err := client.New(kubectlConfig)
	if err != nil {
		return nil, err
	}
	return &KubeClient{
		Client:    cli,
		namespace: kubernetesConfig.Namespace,
	}, nil
}

// WaitForCluster is waiting for at least one node in K8s cluster is ready.
func (k *KubeClient) WaitForCluster(timeout time.Duration) error {
	readyNodesFilterFunc := func() bool {
		nodes, err := k.getReadyNodes()
		if err != nil {
			return false
		}
		return len(nodes) > 0
	}
	return k.kubectlWait(readyNodesFilterFunc, timeout)
}

// WaitForPod is waiting for all pods are up and running.
func (k *KubeClient) WaitForPod(timeout time.Duration) error {
	runningPodsFilterFunc := func() bool {
		runningPods, notRunningPods, err := k.GetPods()
		if err != nil {
			return false
		}
		return len(notRunningPods) == 0 && len(runningPods) > 0
	}
	return k.kubectlWait(runningPodsFilterFunc, timeout)
}

// KubectlWait run K8s request and check results for expected string in a loop every second, unless it expected substring is found or timeout expires.
func (k *KubeClient) kubectlWait(filterFunction func() bool, timeout time.Duration) error {
	requstedTimeout := time.After(timeout)
	for {
		if filterFunction() {
			return nil
		}
		select {
		case <-requstedTimeout:
			return fmt.Errorf("timeout(%s) on K8s call", timeout.String())
		default:
		}
		time.Sleep(1 * time.Second)
	}
}

// GetPods gathers running and not running pods from K8s cluster.
func (k *KubeClient) GetPods() ([]*api.Pod, []*api.Pod, error) {
	pods, err := k.Pods(k.namespace).List(api.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	var runningPods []*api.Pod
	var notRunningPods []*api.Pod

	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case api.PodRunning:
			runningPods = append(runningPods, &pod)
		case api.PodPending:
		case api.PodSucceeded:
		case api.PodFailed:
			notRunningPods = append(notRunningPods, &pod)
		case api.PodUnknown:
			return nil, nil, fmt.Errorf("at least one of pods is in Unknown state")
		}
	}

	return runningPods, notRunningPods, nil
}

func (k *KubeClient) getReadyNodes() ([]*api.Node, error) {
	nodes, err := k.Nodes().List(api.ListOptions{})
	if err != nil {
		return nil, err
	}

	var readyNodes []*api.Node
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != "True" {
				readyNodes = append(readyNodes, &node)
			}
		}
	}

	return readyNodes, nil
}
