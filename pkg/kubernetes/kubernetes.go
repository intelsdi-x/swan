package kubernetes

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
)

const ()

var (
	// path flags contain paths to kubernetes services' binaries. Default values were fetched from
	// systemd units fetched from http://cbs.centos.org/repos/virt7-docker-common-release/x86_64/os/ repo.
	pathKubeAPIServerFlag  = conf.NewFileFlag("kube_apiserver_path", "Path to kube-apiserver binary", "/usr/bin/kube-apiserver")
	pathKubeControllerFlag = conf.NewFileFlag("kube_controller_path", "Path to kube-controller-manager binary", "/usr/bin/kube-controller-manager")
	pathKubeletFlag        = conf.NewFileFlag("kubelet_path", "Path to kubelet binary", "/usr/bin/kubelet")
	pathKubeProxyFlag      = conf.NewFileFlag("kube_proxy_path", "Path to kube-proxy binary", "/usr/bin/kube-proxy")
	pathKubeSchedulerFlag  = conf.NewFileFlag("kube_scheduler_path", "Path to kube-scheduler binary", "/usr/bin/kube-scheduler")
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
	ETCDServers string
	LogLevel    int // 0 is debug.
	KubeAPIPort int
	KubeletPort int
	// Address range to use for services.
	ServiceAddresses string

	// Custom args to kube-apiserver and kubelet.
	KubeAPIArgs   string
	KubeletArgs   string
	KubeProxyArgs string
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
		LogLevel:             0,
		KubeAPIPort:          8080,
		KubeletPort:          10250,
		ServiceAddresses:     "10.254.0.0/16",
	}
}

type kubernetes struct {
	master  executor.Executor
	kubelet executor.Executor
	config  Config
}

// New returns a new Kubernetes cluster instance consists of one master and one kubelet.
// In case of the same executor they will be on the same host (high risk of interferences).
// NOTE: Currently we support only single-kubelet kubernetes.
// We would need to setup flannel/calico for multi-kubelet setup.
func New(master executor.Executor, kubelet executor.Executor, config Config) executor.Launcher {
	return kubernetes{
		master:  master,
		kubelet: kubelet,
		config:  config,
	}
}

// Name returns human readable name for job.
func (m kubernetes) Name() string {
	return "Kubernetes [single-kubelet]"
}

// TODO(bp): Remove that when all command methods will be implemented.
func fakeCommand() string {
	panic("NOT IMPLEMENTED.")
}

// Launch starts the kubernetes cluster. It returns a cluster
// represented as a Task Handle instance.
// Error is returned when Launcher is unable to start a cluster.
// TODO(bp): Add logging.
func (m kubernetes) Launch() (executor.TaskHandle, error) {
	// Launch kube-apiserver using master executor.
	apiHandle, err := m.master.Execute(getKubeAPIServerCommand(m.config))
	if err != nil {
		return nil, fmt.Errorf(
			"Execution of kube-apiserver failed; Command: %s; %s",
			getKubeAPIServerCommand(m.config), err.Error())
	}

	// TODO(bp): Use proper
	controllerHandle, err := m.master.Execute(fakeCommand())
	if err != nil {
		// TODO(bp): Kill already started services.
		return nil, fmt.Errorf(
			"Execution of kube-controller-manager failed; Command: %s; %s",
			fakeCommand(), err.Error())
	}

	schedulerHandle, err := m.master.Execute(fakeCommand())
	if err != nil {
		// TODO(bp): Kill already started services.
		return nil, fmt.Errorf(
			"Execution of kube-scheduler failed; Command: %s; %s",
			fakeCommand(), err.Error())
	}

	// Launch services on minion node.
	proxyHandle, err := m.master.Execute(getKubeProxyCommand(apiHandle, m.config))
	if err != nil {
		// TODO(bp): Kill already started services.
		return nil, fmt.Errorf(
			"Execution of kube-proxy failed; Command: %s; %s",
			getKubeProxyCommand(apiHandle, m.config), err.Error())
	}

	kubeletHandle, err := m.master.Execute(getKubeletCommand(apiHandle, m.config))
	if err != nil {
		// TODO(bp): Kill already started services.
		return nil, fmt.Errorf(
			"Execution of kubelet failed; Command: %s; %s",
			getKubeletCommand(apiHandle, m.config), err.Error())
	}

	// TODO(bp): Maybe we should special handle for that? We don't want to have wait implementation!
	var noWaitHandle executor.TaskHandle
	noWaitHandle = nil
	return executor.NewClusterTaskHandle(noWaitHandle, []executor.TaskHandle{
		apiHandle,
		controllerHandle,
		schedulerHandle,
		proxyHandle,
		kubeletHandle,
	}), nil
}
