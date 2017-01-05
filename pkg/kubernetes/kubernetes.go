package kubernetes

import (
	"fmt"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/utils/netutil"
	"github.com/intelsdi-x/swan/pkg/utils/random"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	serviceListenTimeout = 15 * time.Second

	// waitForReadyNode configuration
	waitForReadyNodeBackOffPeriod = 1 * time.Second
	defaultReadyNodeRetryCount    = 20
	expectedKubelelNodesCount     = 1
)

var (
	// path flags contain paths to kubernetes services' binaries. See README.md for details.
	pathKubeAPIServerFlag   = conf.NewStringFlag("kube_apiserver_path", "Path to kube-apiserver binary", path.Join(fs.GetAthenaBinPath(), "kube-apiserver"))
	pathKubeControllerFlag  = conf.NewStringFlag("kube_controller_path", "Path to kube-controller-manager binary", path.Join(fs.GetAthenaBinPath(), "kube-controller-manager"))
	pathKubeletFlag         = conf.NewStringFlag("kubelet_path", "Path to kubelet binary", path.Join(fs.GetAthenaBinPath(), "kubelet"))
	pathKubeProxyFlag       = conf.NewStringFlag("kube_proxy_path", "Path to kube-proxy binary", path.Join(fs.GetAthenaBinPath(), "kube-proxy"))
	pathKubeSchedulerFlag   = conf.NewStringFlag("kube_scheduler_path", "Path to kube-scheduler binary", path.Join(fs.GetAthenaBinPath(), "kube-scheduler"))
	kubeAPIArgsFlag         = conf.NewStringFlag("kube_apiserver_args", "Additional args for kube-apiserver binary (eg. --admission-control=\"AlwaysAdmit,AddToleration\").", "")
	kubeletArgsFlag         = conf.NewStringFlag("kubelet_args", "Additional args for kubelet binary.", "")
	logLevelFlag            = conf.NewIntFlag("kube_loglevel", "Log level for kubernetes servers", 0)
	allowPrivilegedFlag     = conf.NewBoolFlag("kube_allow_privileged", "Allow containers to request privileged mode on cluster and node level (api server and kubelete ).", false)
	kubeEtcdServersFlag     = conf.NewStringFlag("kube_etcd_servers", "Comma seperated list of etcd servers (full URI: http://ip:port)", "http://127.0.0.1:2379")
	readyNodeRetryCountFlag = conf.NewIntFlag("kube_node_ready_retry_count", "Number of checks that kubelet is ready, before trying setup cluster again (with 1s interval between checks).", defaultReadyNodeRetryCount)
)

// Config contains all data for running kubernetes master & kubelet.
type Config struct {
	PathToKubeAPIServer  string
	PathToKubeController string
	PathToKubeScheduler  string
	PathToKubeProxy      string
	PathToKubelet        string

	// Comma separated list of nodes in the etcd cluster
	EtcdServers        string
	EtcdPrefix         string
	LogLevel           int // 0 is info, 4 - debug (https://github.com/kubernetes/kubernetes/blob/master/docs/devel/logging.md).
	KubeAPIPort        int
	KubeControllerPort int
	KubeSchedulerPort  int
	KubeProxyPort      int
	KubeletPort        int
	AllowPrivileged    bool // Defaults to false.
	// Address range to use for services.
	ServiceAddresses string

	// Custom args to kube-apiserver and kubelet.
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
		PathToKubeAPIServer:  pathKubeAPIServerFlag.Value(),
		PathToKubeController: pathKubeControllerFlag.Value(),
		PathToKubeScheduler:  pathKubeSchedulerFlag.Value(),
		PathToKubeProxy:      pathKubeProxyFlag.Value(),
		PathToKubelet:        pathKubeletFlag.Value(),
		EtcdServers:          kubeEtcdServersFlag.Value(),
		EtcdPrefix:           "/registry",
		LogLevel:             logLevelFlag.Value(),
		AllowPrivileged:      allowPrivilegedFlag.Value(),
		KubeAPIPort:          8080,
		KubeletPort:          10250,
		KubeControllerPort:   10252,
		KubeSchedulerPort:    10251,
		KubeProxyPort:        10249,
		ServiceAddresses:     "10.2.0.0/16",
		KubeletArgs:          kubeletArgsFlag.Value(),
		KubeAPIArgs:          kubeAPIArgsFlag.Value(),
		RetryCount:           2,
	}
}

// UniqueConfig is a constructor for Config with default parameters and random ports and random etcd prefix.
func UniqueConfig() (Config, error) {
	config := DefaultConfig()
	// Create unique etcd prefix to avoid interference with any parallel tests which use same
	// etcd cluster.
	etcdPrefix, err := uuid.NewV4()
	if err != nil {
		return Config{}, errors.Wrap(err, "cannot create random etcd prefix")
	}
	ETCDPrefix := path.Join("/swan/", etcdPrefix.String())
	config.EtcdPrefix = ETCDPrefix

	// NOTE: To reduce the likelihood of port conflict between test kubernetes clusters, we randomly
	// assign a collection of ports to the services. Eventhough previous kubernetes processes
	// have been shut down, ports may be in CLOSE_WAIT state.
	ports := random.Ports(22768, 32768, 5)
	config.KubeAPIPort = ports[0]
	config.KubeletPort = ports[1]
	config.KubeControllerPort = ports[2]
	config.KubeSchedulerPort = ports[3]
	config.KubeProxyPort = ports[4]

	return config, nil
}

// Type used for UT mocking purposes.
type getReadyNodesFunc func(k8sAPIAddress string) ([]api.Node, error)

type kubernetes struct {
	master        executor.Executor
	minion        executor.Executor // Current single minion is strictly connected with getReadyNodes() function and expectedKubelelNodesCount const.
	config        Config
	isListening   netutil.IsListeningFunction // For mocking purposes.
	getReadyNodes getReadyNodesFunc           // For mocking purposes.
}

// New returns a new Kubernetes launcher instance consists of one master and one minion.
// In case of the same executor they will be on the same host (high risk of interferences).
// NOTE: Currently we support only single-kubelet (single-minion) kubernetes. We also does not
// support ip lookup for pods. To support that we need to setup flannel or calico as well. (SCE-551)
func New(master executor.Executor, minion executor.Executor, config Config) executor.Launcher {
	return kubernetes{
		master:        master,
		minion:        minion,
		config:        config,
		isListening:   netutil.IsListening,
		getReadyNodes: getReadyNodes,
	}
}

// Name returns human readable name for job.
func (m kubernetes) Name() string {
	return "Kubernetes [single-kubelet]"
}

// Launch starts the kubernetes cluster. It returns a cluster
// represented as a Task Handle instance.
// Error is returned when Launcher is unable to start a cluster.
func (m kubernetes) Launch() (handle executor.TaskHandle, err error) {
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

func (m kubernetes) tryLaunchCluster() (executor.TaskHandle, error) {
	handle, err := m.launchCluster()
	if err != nil {
		return nil, err
	}

	apiServerAddress := fmt.Sprintf("%s:%d", handle.Address(), m.config.KubeAPIPort)
	err = m.waitForReadyNode(apiServerAddress)
	if err != nil {
		stopClusterErrors := executor.StopCleanAndErase(handle)
		if stopClusterErrors.GetErrIfAny() != nil {
			log.Warningf("Errors while stopping k8s cluster: %v", stopClusterErrors.GetErrIfAny())
		}
		return nil, err
	}

	return handle, nil
}

func (m kubernetes) launchCluster() (executor.TaskHandle, error) {
	// Launch kube-apiserver using master executor.
	apiHandle, err := m.launchService(
		m.master, getKubeAPIServerCommand(m.config), m.config.KubeAPIPort)
	if err != nil {
		return nil, errors.Wrap(err, "cannot launch kube-apiserver using master executor")
	}
	clusterTaskHandle := executor.NewClusterTaskHandle(apiHandle, []executor.TaskHandle{})

	// Launch kube-controller-manager using master executor.
	controllerHandle, err := m.launchService(
		m.master, getKubeControllerCommand(apiHandle, m.config), m.config.KubeControllerPort)
	if err != nil {
		errCol := executor.StopCleanAndErase(clusterTaskHandle)
		errCol.Add(err)
		return nil, errors.Wrap(errCol.GetErrIfAny(), "cannot launch kube-controller-manager using master executor")
	}
	clusterTaskHandle.AddAgent(controllerHandle)

	// Launch kube-scheduler using master executor.
	schedulerHandle, err := m.launchService(
		m.master, getKubeSchedulerCommand(apiHandle, m.config), m.config.KubeSchedulerPort)
	if err != nil {
		errCol := executor.StopCleanAndErase(clusterTaskHandle)
		errCol.Add(err)
		return nil, errors.Wrap(errCol.GetErrIfAny(), "cannot launch kube-scheduler using master executor")
	}
	clusterTaskHandle.AddAgent(schedulerHandle)

	// Launch services on minion node.
	// Launch kube-proxy using minion executor.
	proxyHandle, err := m.launchService(
		m.minion, getKubeProxyCommand(apiHandle, m.config), m.config.KubeProxyPort)
	if err != nil {
		errCol := executor.StopCleanAndErase(clusterTaskHandle)
		errCol.Add(err)
		return nil, errors.Wrap(errCol.GetErrIfAny(), "cannot launch kube-proxy using minion executor")
	}
	clusterTaskHandle.AddAgent(proxyHandle)

	// Launch kubelet using minion executor.
	kubeletHandle, err := m.launchService(
		m.minion, getKubeletCommand(apiHandle, m.config), m.config.KubeletPort)
	if err != nil {
		errCol := executor.StopCleanAndErase(clusterTaskHandle)
		errCol.Add(err)
		return nil, errors.Wrap(errCol.GetErrIfAny(), "cannot launch kubelet using minion executor")
	}
	clusterTaskHandle.AddAgent(kubeletHandle)

	return clusterTaskHandle, err
}

// launchService executes service and check if it is listening on it's endpoint.
func (m kubernetes) launchService(exec executor.Executor, command string, port int) (executor.TaskHandle, error) {
	handle, err := exec.Execute(command)
	if err != nil {
		return nil, errors.Wrapf(err, "execution of command %q on %q failed", command, exec.Name())
	}

	address := fmt.Sprintf("%s:%d", handle.Address(), port)
	if !m.isListening(address, serviceListenTimeout) {
		defer executor.StopCleanAndErase(handle)
		ec, _ := handle.ExitCode()

		return nil, errors.Errorf(
			"failed to connect to service %q on %q: timeout on connection to %q; task status is %v and exit code is %d",
			command, exec.Name(), address, handle.Status(), ec)
	}

	return handle, nil
}

func (m kubernetes) waitForReadyNode(apiServerAddress string) error {
	for idx := 0; idx < readyNodeRetryCountFlag.Value(); idx++ {
		nodes, err := m.getReadyNodes(apiServerAddress)
		if err != nil {
			return err
		}

		if len(nodes) == expectedKubelelNodesCount {
			return nil
		}

		time.Sleep(waitForReadyNodeBackOffPeriod)
	}

	return errors.New("kubelet could not register in time")
}

func getReadyNodes(k8sAPIAddress string) ([]api.Node, error) {
	kubectlConfig := &restclient.Config{
		Host:     k8sAPIAddress,
		Username: "",
		Password: "",
	}

	k8sClient, err := client.New(kubectlConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create new Kubernetes client on %q", k8sAPIAddress)
	}

	nodes, err := k8sClient.Nodes().List(api.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "could not obtain Kubernetes node list on %q", k8sAPIAddress)
	}

	var readyNodes []api.Node
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == api.NodeReady && condition.Status == api.ConditionTrue {
				readyNodes = append(readyNodes, node)
			}
		}
	}

	return readyNodes, nil
}
