package kubernetes

import "fmt"

// getKubeAPIServerCommand returns command for kube-apiserver.
func getKubeAPIServerCommand(config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("apiserver"),
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --allow-privileged=%v", config.AllowPrivileged),
		fmt.Sprintf(" --etcd-servers=%s", config.EtcdServers),
		fmt.Sprintf(" --etcd-prefix=%s", config.EtcdPrefix),
		fmt.Sprintf(" --insecure-bind-address=%s", config.KubeAPIAddr),
		fmt.Sprintf(" --insecure-port=%d", config.KubeAPIPort),
		fmt.Sprintf(" --kubelet-timeout=%s", serviceListenTimeout),
		fmt.Sprintf(" --service-cluster-ip-range=%s", config.ServiceAddresses),
		fmt.Sprintf(" %s", config.KubeAPIArgs),
	)
}

// getKubeControllerCommand returns command for kube-controller-manager.
func getKubeControllerCommand(config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("controller-manager"),
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --master=%s", config.GetKubeAPIAddress()),
		fmt.Sprintf(" --port=%d", config.KubeControllerPort),
		fmt.Sprintf(" %s", config.KubeControllerArgs),
	)
}

// getKubeSchedulerCommand returns command for kube-scheduler.
func getKubeSchedulerCommand(config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("scheduler"),
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --master=%s", config.GetKubeAPIAddress()),
		fmt.Sprintf(" --port=%d", config.KubeSchedulerPort),
		fmt.Sprintf(" %s", config.KubeSchedulerArgs),
	)
}

// getKubeletCommand returns command for kubelet.
func getKubeletCommand(config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("kubelet"),
		fmt.Sprintf(" --allow-privileged=%v", config.AllowPrivileged),
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --port=%d", config.KubeletPort),
		fmt.Sprintf(" --read-only-port=0"),
		fmt.Sprintf(" --api-servers=%s", config.GetKubeAPIAddress()),
		fmt.Sprintf(" %s", config.KubeletArgs),
	)
}

// getKubeProxyCommand returns command for kube-proxy.
func getKubeProxyCommand(config Config) string {
	return fmt.Sprint(
		fmt.Sprintf("proxy"),
		fmt.Sprintf(" --v=%d", config.LogLevel),
		fmt.Sprintf(" --healthz-port=%d", config.KubeProxyPort),
		fmt.Sprintf(" --master=%s", config.GetKubeAPIAddress()),
		fmt.Sprintf(" %s", config.KubeProxyArgs),
	)
}
