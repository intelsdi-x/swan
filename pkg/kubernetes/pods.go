package kubernetes

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

type k8sPodAPI interface {
	killPods(pods []v1.Pod) error
	getPodsFromNode(nodeName string) (result []v1.Pod, err error)
}

type k8sPodAPIImplementation struct {
	client *kubernetes.Clientset
}

func newK8sPodAPI(config Config) k8sPodAPI {
	client, err := kubernetes.NewForConfig(
		&rest.Config{
			Host: fmt.Sprintf("%s:%d", config.KubeAPIAddr, config.KubeAPIPort),
		},
	)
	if err != nil {
		panic(err)
	}
	return &k8sPodAPIImplementation{
		client: client,
	}
}

func (m *k8s) cleanNode(nodeName string, pods []v1.Pod) error {
	err := m.killPods(pods)
	if err != nil {
		return err
	}
	timeout := time.After(30 * time.Second)
	for {
		select {
		case <-timeout:
			log.Errorf("Timeout while cleaning hanging pods on node %q", nodeName)
			return errors.Errorf("timeout while cleaning hanging pods on node %q", nodeName)
		default:
			break
		}
		pods, err := m.getPodsFromNode(nodeName)
		if err != nil {
			return err
		}

		if len(pods) == 0 {
			log.Debugf("Hanging pods on node %q has %d hanging pods running", nodeName, len(pods))
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (m *k8sPodAPIImplementation) getPodsFromNode(nodeName string) (result []v1.Pod, err error) {
	if isLocalhost(nodeName) {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, errors.Wrapf(err, "could not obtain hostname to clean kubernetes node")
		}
		nodeName = hostname
	}
	pods, err := m.client.Pods(v1.NamespaceDefault).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot retrieve pods running on cluster")
	}
	for _, pod := range pods.Items {
		if pod.Spec.NodeName == nodeName {
			result = append(result, pod)
		}
	}
	return result, nil
}

func (m *k8sPodAPIImplementation) killPods(pods []v1.Pod) error {
	podsAPI := m.client.Core().Pods(v1.NamespaceDefault)

	for _, pod := range pods {
		var gracePeriod int64 = 1
		err := podsAPI.Delete(pod.Name, &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
		if err != nil {
			return errors.Wrapf(err, "cannot delete pod")
		}
	}
	return nil
}

func isLocalhost(nodeName string) bool {
	return nodeName == "localhost" || nodeName == "127.0.0.1"
}
