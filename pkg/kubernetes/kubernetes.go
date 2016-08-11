package kubernetes

import (
	"fmt"
	"path"
	"time"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/utils/netutil"
	"github.com/pkg/errors"
)

const serviceListenTimeout = 5 * time.Second

var (
	// path flags contain paths to kubernetes services' binaries. See README.md for details.
	pathKubeAPIServerFlag  = conf.NewFileFlag("kube_apiserver_path", "Path to kube-apiserver binary", path.Join(fs.GetSwanBinPath(), "kube-apiserver"))
	pathKubeControllerFlag = conf.NewFileFlag("kube_controller_path", "Path to kube-controller-manager binary", path.Join(fs.GetSwanBinPath(), "kube-controller-manager"))
	pathKubeletFlag        = conf.NewFileFlag("kubelet_path", "Path to kubelet binary", path.Join(fs.GetSwanBinPath(), "kubelet"))
	pathKubeProxyFlag      = conf.NewFileFlag("kube_proxy_path", "Path to kube-proxy binary", path.Join(fs.GetSwanBinPath(), "kube-proxy"))
	pathKubeSchedulerFlag  = conf.NewFileFlag("kube_scheduler_path", "Path to kube-scheduler binary", path.Join(fs.GetSwanBinPath(), "kube-scheduler"))
	kubeletArgsFlag        = conf.NewStringFlag("kubelet_args", "Additional args for kubelet binary.", "")
	logLevelFlag           = conf.NewIntFlag("kube_loglevel", "Log level for kubernetes servers", 0)
)

// Config contains all data for running kubernetes master & kubelet.
type Config struct {
	PathToKubeAPIServer  string
	PathToKubeController string
	PathToKubeScheduler  string
	PathToKubeProxy      string
	PathToKubelet        string

	// TODO(bp): Consider exposing these via flags (SCE-547)
	// Comma separated list of nodes in the etcd cluster
	ETCDServers        string
	LogLevel           int // 0 is debug.
	KubeAPIPort        int
	KubeControllerPort int
	KubeSchedulerPort  int
	KubeProxyPort      int
	KubeletPort        int
	// Address range to use for services.
	ServiceAddresses string

	// Custom args to kube-apiserver and kubelet.
	KubeAPIArgs        string
	KubeControllerArgs string
	KubeSchedulerArgs  string
	KubeletArgs        string
	KubeProxyArgs      string
}

// DefaultConfig is a constructor for Config with default parameters.
func DefaultConfig() Config {
	return Config{
		PathToKubeAPIServer:  pathKubeAPIServerFlag.Value(),
		PathToKubeController: pathKubeControllerFlag.Value(),
		PathToKubeScheduler:  pathKubeSchedulerFlag.Value(),
		PathToKubeProxy:      pathKubeProxyFlag.Value(),
		PathToKubelet:        pathKubeletFlag.Value(),
		ETCDServers:          "http://127.0.0.1:2379",
		LogLevel:             logLevelFlag.Value(),
		KubeAPIPort:          8080,
		KubeletPort:          10250,
		KubeControllerPort:   10252,
		KubeSchedulerPort:    10251,
		KubeProxyPort:        10249,
		ServiceAddresses:     "10.2.0.0/16",
		KubeletArgs:          kubeletArgsFlag.Value(),
	}
}

type kubernetes struct {
	master      executor.Executor
	minion      executor.Executor
	config      Config
	isListening netutil.IsListeningFunction // For mocking purposes.
}

// New returns a new Kubernetes launcher instance consists of one master and one minion.
// In case of the same executor they will be on the same host (high risk of interferences).
// NOTE: Currently we support only single-kubelet (single-minion) kubernetes. We also does not
// support ip lookup for pods. To support that we need to setup flannel or calico as well. (SCE-551)
func New(master executor.Executor, minion executor.Executor, config Config) executor.Launcher {
	return kubernetes{
		master:      master,
		minion:      minion,
		config:      config,
		isListening: netutil.IsListening,
	}
}

// Name returns human readable name for job.
func (m kubernetes) Name() string {
	return "Kubernetes [single-kubelet]"
}

// launchService executes service and check if it is listening on it's endpoint.
func (m kubernetes) launchService(exec executor.Executor, command string, port int) (executor.TaskHandle, error) {
	handle, err := exec.Execute(command)
	if err != nil {
		return nil, errors.Wrapf(err, "execution of command %q on %q failed", command, exec.Name())
	}

	address := fmt.Sprintf("%s:%d", handle.Address(), port)
	if !m.isListening(address, serviceListenTimeout) {
		executor.LogUnsucessfulExecution(command, exec.Name(), handle)

		defer handle.EraseOutput()
		defer handle.Clean()
		defer handle.Stop()

		return nil, errors.Errorf(
			"failed to connect to service %q on %q: timeout on connection to %q",
			command, exec.Name(), address)
	}

	return handle, nil
}

func (m kubernetes) stopAndCleanupCluster(clusterTaskHandle executor.ClusterTaskHandle) errcollection.ErrorCollection {
	if clusterTaskHandle == nil {
		return nil
	}

	var errorCollection errcollection.ErrorCollection

	errorCollection.Add(clusterTaskHandle.Stop())
	errorCollection.Add(clusterTaskHandle.Clean())
	errorCollection.Add(clusterTaskHandle.EraseOutput())

	return errorCollection
}

// Launch starts the kubernetes cluster. It returns a cluster
// represented as a Task Handle instance.
// Error is returned when Launcher is unable to start a cluster.
func (m kubernetes) Launch() (executor.TaskHandle, error) {
	// Launch kube-apiserver using master executor.
	apiHandle, err := m.launchService(
		m.master, getKubeAPIServerCommand(m.config), m.config.KubeAPIPort)
	if err != nil {
		return nil, err
	}
	clusterTaskHandle := executor.NewClusterTaskHandle(apiHandle, []executor.TaskHandle{})

	// Launch kube-controller-manager using master executor.
	controllerHandle, err := m.launchService(
		m.master, getKubeControllerCommand(apiHandle, m.config), m.config.KubeControllerPort)
	if err != nil {
		errCol := m.stopAndCleanupCluster(clusterTaskHandle)
		errCol.Add(err)
		return nil, errCol.GetErrIfAny()
	}
	clusterTaskHandle.AddAgent(controllerHandle)

	// Launch kube-scheduler using master executor.
	schedulerHandle, err := m.launchService(
		m.master, getKubeSchedulerCommand(apiHandle, m.config), m.config.KubeSchedulerPort)
	if err != nil {
		errCol := m.stopAndCleanupCluster(clusterTaskHandle)
		errCol.Add(err)
		return nil, errCol.GetErrIfAny()
	}
	clusterTaskHandle.AddAgent(schedulerHandle)

	// Launch services on minion node.
	// Launch kube-proxy using minion executor.
	proxyHandle, err := m.launchService(
		m.minion, getKubeProxyCommand(apiHandle, m.config), m.config.KubeProxyPort)
	if err != nil {
		errCol := m.stopAndCleanupCluster(clusterTaskHandle)
		errCol.Add(err)
		return nil, errCol.GetErrIfAny()
	}
	clusterTaskHandle.AddAgent(proxyHandle)

	// Launch kubelet using minion executor.
	kubeletHandle, err := m.launchService(
		m.minion, getKubeletCommand(apiHandle, m.config), m.config.KubeletPort)
	if err != nil {
		errCol := m.stopAndCleanupCluster(clusterTaskHandle)
		errCol.Add(err)
		return nil, errCol.GetErrIfAny()
	}
	clusterTaskHandle.AddAgent(kubeletHandle)

	// NOTE: We may add a simple pre-health-check here for instantiating one pod or checking how
	// many nodes we have in cluster. (SCE-548)

	return clusterTaskHandle, nil
}
