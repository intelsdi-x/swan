# Kubernetes cluster launcher

This launcher starts the kubernetes cluster. It returns a cluster represented as a Task Handle instance.
You can specify two executors:
- One executor specify how to execute master services (and on what host as well) like `apiserver`, `controller-manager` and `scheduler`.
- Second are for minion services  like `kubelet` and `proxy`

## Prerequisites

- install etcd (`yum install -y etcd`), docker and iptables.
- Download k8s binaries (current launcher tested on `1.5` k8s) from
e.g [here](https://github.com/kubernetes/kubernetes/releases)

## Note:

It is recommended to use 2 machines with CentOS. It is important for
swan experiment to have k8s minion not being interfered by master services.
