package kubernetes

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/rest"
)

func getReadyNodes(k8sAPIAddress string) ([]v1.Node, error) {
	kubectlConfig := &rest.Config{
		Host:     k8sAPIAddress,
		Username: "",
		Password: "",
	}

	k8sClientset, err := kubernetes.NewForConfig(kubectlConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create XXXX new Kubernetes client on %q", k8sAPIAddress)
	}

	nodes, err := k8sClientset.Core().Nodes().List(api.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "could not obtain Kubernetes node list on %q", k8sAPIAddress)
	}

	var readyNodes []v1.Node
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				readyNodes = append(readyNodes, node)
			}
		}
	}

	return readyNodes, nil
}
