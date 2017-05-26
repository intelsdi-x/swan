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

package kubernetes

import (
	"fmt"
	"path"
	"time"

	"k8s.io/client-go/pkg/api/v1"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/utils/netutil"
	"github.com/intelsdi-x/swan/pkg/utils/random"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"github.com/pkg/errors"
)

const (
	serviceListenTimeout = 15 * time.Second

	// waitForReadyNode configuration
	waitForReadyNodeBackOffPeriod = 1 * time.Second
	nodeCheckRetryCount           = 20
	expectedKubeletNodesCount     = 1
)

var (
	kubeEtcdServersFlag = conf.NewStringFlag("kubernetes_cluster_etcd_servers", "Comma seperated list of etcd servers (full URI: http://ip:port)", "http://127.0.0.1:2379")

	//KubernetesMasterFlag indicates where Kubernetes control plane will be launched.
	KubernetesMasterFlag = conf.NewStringFlag("kubernetes_cluster_run_control_plane_on_host", "Address of a host where Kubernetes control plane will be run (when using -kubernetes and not connecting to existing cluster).", "127.0.0.1")

	kubeCleanLeftPods = conf.NewBoolFlag("kubernetes_cluster_clean_left_pods_on_startup", "Delete all pods which are detected during cluster startup. Usefull after dirty shutdown when some pods may not be properly deleted.", false)
)

type kubeCommand struct {
	exec            executor.Executor
	raw             string
	healthCheckPort int
}

// Config contains all data for running kubernetes master & kubelet.
type Config struct {
	// Comma separated list of nodes in the etcd cluster
	EtcdServers        string
	EtcdPrefix         string
	LogLevel           int // 0 is info, 4 - debug (https://github.com/kubernetes/kubernetes/blob/master/docs/devel/logging.md).
	KubeAPIAddr        string
	KubeAPIPort        int
	KubeControllerPort int
	KubeSchedulerPort  int
	KubeProxyPort      int
	KubeletPort        int
	AllowPrivileged    bool
	// Address range to use for services.
	ServiceAddresses string

	// Optional configuration option for cleaning
	KubeletHost string

	// Custom args to apiserver and kubelet.
	KubeAPIArgs        string
	KubeControllerArgs string
	KubeSchedulerArgs  string
	KubeletArgs        string
	KubeProxyArgs      string

	// Launcher configuration
	RetryCount uint64
}

// DefaultConfig is a constructor for Config with default parameters.
func DefaultConfig() Config {
	return Config{
		EtcdServers:        kubeEtcdServersFlag.Value(),
		EtcdPrefix:         "/swan",
		LogLevel:           0,
		AllowPrivileged:    true,
		KubeAPIAddr:        KubernetesMasterFlag.Value(), // TODO(skonefal): This should not be part of config.
		KubeAPIPort:        8080,
		KubeletPort:        10250,
		KubeControllerPort: 10252,
		KubeSchedulerPort:  10251,
		KubeProxyPort:      10249,
		ServiceAddresses:   "10.2.0.0/16",
		RetryCount:         0,
	}
}

// GetKubeAPIAddress returns kube api server in HTTP format.
func (c *Config) GetKubeAPIAddress() string {
	return fmt.Sprintf("http://%s:%d", c.KubeAPIAddr, c.KubeAPIPort)
}

// UniqueConfig is a constructor for Config with default parameters and random ports and random etcd prefix.
func UniqueConfig() Config {
	config := DefaultConfig()
	// Create unique etcd prefix to avoid interference with any parallel tests which use same
	// etcd cluster.
	config.EtcdPrefix = path.Join("/swan/", uuid.New())

	// NOTE: To reduce the likelihood of port conflict between test kubernetes clusters, we randomly
	// assign a collection of ports to the services. Eventhough previous kubernetes processes
	// have been shut down, ports may be in CLOSE_WAIT state.
	ports := random.Ports(5)
	config.KubeAPIPort = ports[0]
	config.KubeletPort = ports[1]
	config.KubeControllerPort = ports[2]
	config.KubeSchedulerPort = ports[3]
	config.KubeProxyPort = ports[4]

	return config
}

// Type used for UT mocking purposes.
type getReadyNodesFunc func(k8sAPIAddress string) ([]v1.Node, error)

type k8s struct {
	master executor.Executor
	minion executor.Executor // Current single minion is strictly connected with getReadyNodes() function and expectedKubeletNodesCount const.
	config Config

	k8sPodAPI // Private interface

	isListening   netutil.IsListeningFunction // For mocking purposes.
	getReadyNodes getReadyNodesFunc           // For mocking purposes.

	kubeletHost string // Filled by Kubelet TaskHandle
}

// New returns a new Kubernetes launcher instance consists of one master and one minion.
// In case of the same executor they will be on the same host (high risk of interferences).
// NOTE: Currently we support only single-kubelet (single-minion) kubernetes.
func New(master executor.Executor, minion executor.Executor, config Config) executor.Launcher {

	return &k8s{
		master:        master,
		minion:        minion,
		config:        config,
		k8sPodAPI:     newK8sPodAPI(config),
		isListening:   netutil.IsListening,
		getReadyNodes: getReadyNodes,
	}
}

// Name returns human readable name for job.
func (m *k8s) Name() string {
	return "Kubernetes [single-kubelet]"
}

// Launch starts the kubernetes cluster. It returns a cluster
// represented as a Task Handle instance.
// Error is returned when Launcher is unable to start a cluster.
func (m *k8s) Launch() (handle executor.TaskHandle, err error) {
	for retry := uint64(0); retry <= m.config.RetryCount; retry++ {
		handle, err = m.tryLaunchCluster()
		if err != nil {
			log.Warningf("could not launch Kubernetes cluster: %q. Retry number: %d", err.Error(), retry)
			continue
		}
		return handle, nil
	}
	log.Errorf("Could not launch Kubernetes cluster: %q", err.Error())
	return nil, err
}

func (m *k8s) tryLaunchCluster() (executor.TaskHandle, error) {
	handle, err := m.launchCluster()
	if err != nil {
		return nil, err
	}

	apiServerAddress := fmt.Sprintf("%s:%d", handle.Address(), m.config.KubeAPIPort)
	err = m.waitForReadyNode(apiServerAddress)
	if err != nil {
		stopErr := handle.Stop()
		if stopErr != nil {
			log.Warningf("Errors while stopping k8s cluster: %v", stopErr)
		}
		return nil, err
	}
	// Optional removal of the unwanted pods in swan's namespace
	pods, err := m.getPodsFromNode(m.kubeletHost)
	if err != nil {
		log.Warningf("Could not retreive list of pods from host %s. Error: %s", m.kubeletHost, err)
		// if getPodsFromNode returns error it means cluster is not useable. Delete it.
		stopErr := handle.Stop()
		if stopErr != nil {
			log.Errorf("Errors while stopping k8s cluster: %v", stopErr)
		}
		return nil, err
	}
	if len(pods) != 0 {
		if kubeCleanLeftPods.Value() {
			log.Infof("Kubelet on node %q has %d dangling pods. Attempt to clean them", m.kubeletHost, len(pods))
			err = m.cleanNode(m.kubeletHost, pods)
			if err != nil {
				log.Errorf("Could not clean dangling pods: %s", err)
			} else {
				log.Infof("Dangling pods on node %q has been deleted", m.kubeletHost)
			}
		} else {
			stopErr := handle.Stop()
			if stopErr != nil {
				log.Errorf("Errors while stopping k8s cluster: %v", stopErr)
			}
			return nil, errors.Errorf("Kubelet on node %q has %d dangling pods. Use `kubectl` to delete them or set %q flag to let Swan remove them", m.kubeletHost, len(pods), kubeCleanLeftPods.Name)
		}
	}
	return handle, nil
}

func (m *k8s) launchCluster() (executor.TaskHandle, error) {
	// Launch apiserver using master executor.
	kubeAPIServer := m.getKubeAPIServerCommand()
	apiHandle, err := m.launchService(kubeAPIServer)
	if err != nil {
		return nil, errors.Wrap(err, "cannot launch apiserver using master executor")
	}
	clusterTaskHandle := executor.NewClusterTaskHandle(apiHandle, []executor.TaskHandle{})

	// Launch controller-manager using master executor.
	kubeController := m.getKubeControllerCommand()
	controllerHandle, err := m.launchService(kubeController)
	if err != nil {
		var errCol errcollection.ErrorCollection
		errCol.Add(clusterTaskHandle.Stop())
		errCol.Add(err)
		return nil, errors.Wrap(errCol.GetErrIfAny(), "cannot launch controller-manager using master executor")
	}
	clusterTaskHandle.AddAgent(controllerHandle)

	// Launch scheduler using master executor.
	kubeScheduler := m.getKubeSchedulerCommand()
	schedulerHandle, err := m.launchService(kubeScheduler)
	if err != nil {
		var errCol errcollection.ErrorCollection
		errCol.Add(clusterTaskHandle.Stop())
		errCol.Add(err)
		return nil, errors.Wrap(errCol.GetErrIfAny(), "cannot launch scheduler using master executor")
	}
	clusterTaskHandle.AddAgent(schedulerHandle)

	// Launch services on minion node.
	// Launch proxy using minion executor.
	kubeProxyCommand := m.getKubeProxyCommand()
	proxyHandle, err := m.launchService(kubeProxyCommand)
	if err != nil {
		var errCol errcollection.ErrorCollection
		errCol.Add(clusterTaskHandle.Stop())
		errCol.Add(err)
		return nil, errors.Wrap(errCol.GetErrIfAny(), "cannot launch proxy using minion executor")
	}
	clusterTaskHandle.AddAgent(proxyHandle)

	// Launch kubelet using minion executor.
	kubeletCommand := m.getKubeletCommand()
	kubeletHandle, err := m.launchService(kubeletCommand)
	if err != nil {
		var errCol errcollection.ErrorCollection
		errCol.Add(clusterTaskHandle.Stop())
		errCol.Add(err)
		return nil, errors.Wrap(errCol.GetErrIfAny(), "cannot launch kubelet using minion executor")
	}
	clusterTaskHandle.AddAgent(kubeletHandle)
	m.kubeletHost = kubeletHandle.Address()

	return clusterTaskHandle, err
}

// launchService executes service and check if it is listening on it's endpoint.
func (m *k8s) launchService(command kubeCommand) (executor.TaskHandle, error) {
	handle, err := command.exec.Execute(command.raw)
	if err != nil {
		return nil, errors.Wrapf(err, "execution of command %q on %q failed", command.raw, command.exec.Name())
	}

	address := fmt.Sprintf("%s:%d", handle.Address(), command.healthCheckPort)
	if !m.isListening(address, serviceListenTimeout) {
		defer handle.Stop()
		ec, _ := handle.ExitCode()

		return nil, errors.Errorf(
			"failed to connect to service %q on %q: timeout on connection to %q; task status is %v and exit code is %d",
			command.raw, command.exec.Name(), address, handle.Status(), ec)
	}

	return handle, nil
}

// getKubeAPIServerCommand returns command for apiserver.
func (m *k8s) getKubeAPIServerCommand() kubeCommand {
	return kubeCommand{m.master,
		fmt.Sprint(
			fmt.Sprintf("hyperkube apiserver"),
			fmt.Sprintf(" --v=%d", m.config.LogLevel),
			fmt.Sprintf(" --allow-privileged=%v", m.config.AllowPrivileged),
			fmt.Sprintf(" --etcd-servers=%s", m.config.EtcdServers),
			fmt.Sprintf(" --etcd-prefix=%s", m.config.EtcdPrefix),
			fmt.Sprintf(" --insecure-bind-address=%s", m.config.KubeAPIAddr),
			fmt.Sprintf(" --insecure-port=%d", m.config.KubeAPIPort),
			fmt.Sprintf(" --secure-port 0"),
			fmt.Sprintf(" --kubelet-timeout=%s", serviceListenTimeout),
			fmt.Sprintf(" --service-cluster-ip-range=%s", m.config.ServiceAddresses),
			fmt.Sprintf(" --advertise-address=%s", m.config.KubeAPIAddr),
			fmt.Sprintf(" --cert-dir=/tmp/kubernetes"),
			fmt.Sprintf(" %s", m.config.KubeAPIArgs),
		), m.config.KubeAPIPort}
}

// getKubeControllerCommand returns command for controller-manager.
func (m *k8s) getKubeControllerCommand() kubeCommand {
	return kubeCommand{m.master,
		fmt.Sprint(
			fmt.Sprintf("hyperkube controller-manager"),
			fmt.Sprintf(" --v=%d", m.config.LogLevel),
			fmt.Sprintf(" --master=%s", m.config.GetKubeAPIAddress()),
			fmt.Sprintf(" --port=%d", m.config.KubeControllerPort),
			fmt.Sprintf(" %s", m.config.KubeControllerArgs),
		), m.config.KubeControllerPort}
}

// getKubeSchedulerCommand returns command for scheduler.
func (m *k8s) getKubeSchedulerCommand() kubeCommand {
	return kubeCommand{m.master,
		fmt.Sprint(
			fmt.Sprintf("hyperkube scheduler"),
			fmt.Sprintf(" --v=%d", m.config.LogLevel),
			fmt.Sprintf(" --master=%s", m.config.GetKubeAPIAddress()),
			fmt.Sprintf(" --port=%d", m.config.KubeSchedulerPort),
			fmt.Sprintf(" %s", m.config.KubeSchedulerArgs),
		), m.config.KubeSchedulerPort}
}

// getKubeletCommand returns command for kubelet.
func (m *k8s) getKubeletCommand() kubeCommand {
	return kubeCommand{m.minion,
		fmt.Sprint(
			fmt.Sprintf("hyperkube kubelet"),
			fmt.Sprintf(" --allow-privileged=%v", m.config.AllowPrivileged),
			fmt.Sprintf(" --v=%d", m.config.LogLevel),
			fmt.Sprintf(" --port=%d", m.config.KubeletPort),
			fmt.Sprintf(" --read-only-port=0"),
			fmt.Sprintf(" --api-servers=%s", m.config.GetKubeAPIAddress()),
			fmt.Sprintf(" %s", m.config.KubeletArgs),
		), m.config.KubeletPort}
}

// getKubeProxyCommand returns command for proxy.
func (m *k8s) getKubeProxyCommand() kubeCommand {
	return kubeCommand{m.minion,
		fmt.Sprint(
			fmt.Sprintf("hyperkube proxy"),
			fmt.Sprintf(" --v=%d", m.config.LogLevel),
			fmt.Sprintf(" --healthz-port=%d", m.config.KubeProxyPort),
			fmt.Sprintf(" --master=%s", m.config.GetKubeAPIAddress()),
			fmt.Sprintf(" %s", m.config.KubeProxyArgs),
		), m.config.KubeProxyPort}
}

func (m *k8s) waitForReadyNode(apiServerAddress string) error {
	for idx := 0; idx < nodeCheckRetryCount; idx++ {
		nodes, err := m.getReadyNodes(apiServerAddress)
		if err != nil {
			return err
		}

		if len(nodes) == expectedKubeletNodesCount {
			return nil
		}

		time.Sleep(waitForReadyNodeBackOffPeriod)
	}

	return errors.New("kubelet could not register in time")
}
