package kubernetes

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/http"
	"github.com/pkg/errors"
	"net"
	"time"
)

var (
	// path flags contain paths to kubernetes services' binaries. Default values were fetched from
	// systemd units fetched from http://cbs.centos.org/repos/virt7-docker-common-release/x86_64/os/ repo.
	pathKubeAPIServerFlag  = conf.NewFileFlag("kube_apiserver_path", "Path to kube-apiserver binary", "/usr/bin/kube-apiserver")
	pathKubeControllerFlag = conf.NewFileFlag("kube_controller_path", "Path to kube-controller-manager binary", "/usr/bin/kube-controller-manager")
	pathKubeletFlag        = conf.NewFileFlag("kubelet_path", "Path to kubelet binary", "/usr/bin/kubelet")
	pathKubeProxyFlag      = conf.NewFileFlag("kube_proxy_path", "Path to kube-proxy binary", "/usr/bin/kube-proxy")
	pathKubeSchedulerFlag  = conf.NewFileFlag("kube_scheduler_path", "Path to kube-scheduler binary", "/usr/bin/kube-scheduler")
	pathKubectlFlag        = conf.NewFileFlag("kubectl_path", "Path to kubectl binary", "/usr/bin/kubectl")
	logLevelFlag           = conf.NewIntFlag("kube_loglevel", "Log level for kubernetes servers", 0)
)

// Config contains all data for running kubernetes master & kubelet.
type Config struct {
	PathToKubeAPIServer  string
	PathToKubeController string
	PathToKubeScheduler  string
	PathToKubeProxy      string
	PathToKubelet        string

	// TODO(bp): Expose these via flags.
	// Comma separated list of nodes in the etcd cluster
	ETCDServers        string
	LogLevel           int // 0 is debug.
	KubeAPIPort        int
	KubeControllerPort int
	KubeSchedulerPort  int
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
		ServiceAddresses:     "10.254.0.0/16",
	}
}

type kubernetes struct {
	master      executor.Executor
	minion      executor.Executor
	config      Config
	isListening http.IsListeningFunction // For mocking purposes.
}

// New returns a new Kubernetes launcher instance consists of one master and one minion.
// In case of the same executor they will be on the same host (high risk of interferences).
// NOTE: Currently we support only single-kubelet (single-minion) kubernetes.
// We would need to setup flannel/calico for multi-kubelet setup.
func New(master executor.Executor, minion executor.Executor, config Config) executor.Launcher {
	return kubernetes{
		master: master,
		minion: minion,
		config: config,
	}
}

// Name returns human readable name for job.
func (m kubernetes) Name() string {
	return "Kubernetes [single-kubelet]"
}

func (m kubernetes) launchKubeAPI() (executor.TaskHandle, error) {
	// Launch kube-apiserver using master executor.
	apiHandle, err := m.master.Execute(getKubeAPIServerCommand(m.config))
	if err != nil {
		return nil, errors.Wrapf(err,
			"Execution of kube-apiserver failed; Command: %s",
			getKubeAPIServerCommand(m.config))
	}

	address := fmt.Sprintf("%s:%d", apiHandle.Address(), m.config.KubeAPIPort)
	if !m.tryConnect(address, 5*time.Second) {
		file, fileErr := apiHandle.StderrFile()
		details := ""
		if fileErr == nil {
			// TODO(bp): I would suggest to implement helper for catting last three
			// lines of this output (when some flag will be specified to do so).
			details = fmt.Sprintf(" Check %s file for details", file.Name())
		}

		return nil, errors.Errorf("Failed to connect to kube-apiserver instance. "+
			"Timeout on connection to %q.%s", address, details)
	}

	return apiHandle, nil
}

// Launch starts the kubernetes cluster. It returns a cluster
// represented as a Task Handle instance.
// Error is returned when Launcher is unable to start a cluster.
func (m kubernetes) Launch() (executor.TaskHandle, error) {
	apiHandle, err := m.launchKubeAPI()
	if err != nil {
		return nil, err
	}
	clusterTaskHandle := executor.NewClusterTaskHandle(apiHandle, []executor.TaskHandle{})

	// DEBUG
	apiHandle.Wait(0 * time.Nanosecond)

	// Launch kube-controller-manager using master executor.
	controllerHandle, err := m.master.Execute(getKubeControllerCommand(apiHandle, m.config))
	if err != nil {
		clusterTaskHandle.Stop()
		clusterTaskHandle.Clean()
		// TODO(bp): Erase output as well?
		return nil, errors.Wrapf(err,
			"Execution of kube-controller-manager failed; Command: %s",
			getKubeControllerCommand(apiHandle, m.config))
	}
	clusterTaskHandle.AddAgent(controllerHandle)

	// Launch kube-scheduler using master executor.
	schedulerHandle, err := m.master.Execute(getKubeSchedulerCommand(apiHandle, m.config))
	if err != nil {
		clusterTaskHandle.Stop()
		clusterTaskHandle.Clean()
		// TODO(bp): Erase output as well?
		return nil, errors.Wrapf(err,
			"Execution of kube-scheduler failed; Command: %s",
			getKubeSchedulerCommand(apiHandle, m.config))
	}
	clusterTaskHandle.AddAgent(schedulerHandle)

	// Launch services on minion node.
	// Launch kube-proxy using minion executor.
	proxyHandle, err := m.minion.Execute(getKubeProxyCommand(apiHandle, m.config))
	if err != nil {
		clusterTaskHandle.Stop()
		clusterTaskHandle.Clean()
		// TODO(bp): Erase output as well?
		return nil, errors.Wrapf(err,
			"Execution of kube-proxy failed; Command: %s",
			getKubeProxyCommand(apiHandle, m.config))
	}
	clusterTaskHandle.AddAgent(proxyHandle)

	// Launch kubelet using minion executor.
	kubeletHandle, err := m.minion.Execute(getKubeletCommand(apiHandle, m.config))
	if err != nil {
		clusterTaskHandle.Stop()
		clusterTaskHandle.Clean()
		// TODO(bp): Erase output as well?
		return nil, errors.Wrapf(err,
			"Execution of kubelet failed; Command: %s",
			getKubeletCommand(apiHandle, m.config))
	}
	clusterTaskHandle.AddAgent(kubeletHandle)

	return clusterTaskHandle, nil
}
