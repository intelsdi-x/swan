package kubernetes

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
)

// getKubeAPIServerCommand returns command for kube-apiserver.
func getKubeAPIServerCommand(config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("%s", config.PathToKubeAPIServer),
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --allow-privileged=false"), // Privileged containers are not allowed.
		fmt.Sprintf(" --etcd-servers=%s", config.ETCDServers),
		fmt.Sprintf(" --etcd-prefix=%s",  config.ETCDPrefix),
		fmt.Sprintf(" --insecure-bind-address=0.0.0.0"),
		fmt.Sprintf(" --insecure-port=%d", config.KubeAPIPort),
		fmt.Sprintf(" --kubelet-timeout=5s"), // Default for now.
		fmt.Sprintf(" --service-cluster-ip-range=%s", config.ServiceAddresses),
		fmt.Sprintf(" %s", config.KubeAPIArgs),
	)
}

// getKubeControllerCommand returns command for kube-controller-manager.
func getKubeControllerCommand(kubeAPIAddr executor.TaskHandle, config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("%s", config.PathToKubeController),
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --address=0.0.0.0"),
		fmt.Sprintf(" --master=http://%s:%d", kubeAPIAddr.Address(), config.KubeAPIPort),
		fmt.Sprintf(" --port=%d", config.KubeControllerPort),
		fmt.Sprintf(" %s", config.KubeControllerArgs),
	)
}

// getKubeSchedulerCommand returns command for kube-scheduler.
func getKubeSchedulerCommand(kubeAPIAddr executor.TaskHandle, config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("%s", config.PathToKubeScheduler),
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --address=0.0.0.0"),
		fmt.Sprintf(" --master=http://%s:%d", kubeAPIAddr.Address(), config.KubeAPIPort),
		fmt.Sprintf(" --port=%d", config.KubeSchedulerPort),
		fmt.Sprintf(" %s", config.KubeSchedulerArgs),
	)
}

// getKubeletCommand returns command for kubelet.
func getKubeletCommand(kubeAPIAddr executor.TaskHandle, config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("%s", config.PathToKubelet),
		fmt.Sprintf(" --allow-privileged=false"), // Privileged containers are not allowed.
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --address=0.0.0.0"),
		fmt.Sprintf(" --port=%d", config.KubeletPort),
		fmt.Sprintf(" --read-only-port=0"),
		fmt.Sprintf(" --api-servers=http://%s:%d", kubeAPIAddr.Address(), config.KubeAPIPort),
		fmt.Sprintf(" %s", config.KubeletArgs),
	)
}

// getKubeProxyCommand returns command for kube-proxy.
func getKubeProxyCommand(kubeAPIAddr executor.TaskHandle, config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("%s", config.PathToKubeProxy),
		fmt.Sprintf(" --bind-address=0.0.0.0"),
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --healthz-port=%d", config.KubeProxyPort),
		fmt.Sprintf(" --master=http://%s:%d", kubeAPIAddr.Address(), config.KubeAPIPort),
		fmt.Sprintf(" %s", config.KubeProxyArgs),
	)
}
