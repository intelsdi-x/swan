package kubernetes

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
)

func getCommonOptions(config Config) string {
	return fmt.Sprint(
		fmt.Sprintf(" --etcd-servers=%s", config.ETCDServers),
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --allow-privileged=false"),
	)
}

// getKubeAPIServerCommand returns command for kube-apiserver.
func getKubeAPIServerCommand(config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("%s%s", config.PathToKubeAPIServer, getCommonOptions(config)),
		// TODO(bp): Test it. Documentation vs help description is conflicting.
		fmt.Sprintf(" --address=0.0.0.0"),
		// TODO(bp): Test it. Documentation vs help description is conflicting.
		fmt.Sprintf(" --port=%d", config.KubeAPIPort),
		fmt.Sprintf(" --kubelet-port=%d", config.KubeletPort),
		fmt.Sprintf(" --service-cluster-ip-range=%s", config.ServiceAddresses),
		fmt.Sprintf(" %s", config.KubeAPIArgs),
	)
}

// TODO(bp): Implement command getter for kube-controller-manager & kube-scheduler as well.

// getKubeletCommand returns command for kubelet.
func getKubeletCommand(kubeAPIAddr executor.TaskHandle, config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("%s%s", config.PathToKubelet, getCommonOptions(config)),
		fmt.Sprintf(" --address=0.0.0.0"),
		fmt.Sprintf(" --port=%d", config.KubeletPort),
		fmt.Sprintf(" --api-servers=http://%s:%d", kubeAPIAddr.Address(), config.KubeletPort),
		fmt.Sprintf(" %s", config.KubeletArgs),
	)
}

// getKubeProxyCommand returns command for kube-proxy.
func getKubeProxyCommand(kubeAPIAddr executor.TaskHandle, config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("%s%s", config.PathToKubeProxy, getCommonOptions(config)),
		fmt.Sprintf(" --api-servers=http://%s:%d", kubeAPIAddr.Address(), config.KubeletPort),
		fmt.Sprintf(" %s", config.KubeProxyArgs),
	)
}
