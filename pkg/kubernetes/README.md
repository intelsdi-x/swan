# Kubernetes cluster launcher

This launcher starts the kubernetes cluster. It returns a cluster represented as a Task Handle instance.
You can specify two executors:
- One executor specify how to execute master services (and on what host as well) like `kube-apiserver`, `kube-controller-manager` and `kube-scheduler`.
- Second are for minion services  like `kubelet` and `kube-proxy`

## Prerequisites

- 2 machines with CentOS (to have k8s minion not being interfered by master services)
- install etcd (`yum install -y etcd`), docker and iptables.
- Download k8s binaries (current launcher tested on `1.3` k8s) from
e.g [here](https://github.com/kubernetes/kubernetes/releases)