# Swan Flags

Most important flags for running Sensitivity Profile are listed here. Full list of flags is shown through `memcached-sensitivity-profile -help` command, or listed in [Example Config Dump](config_dump_example.md).   

## Swan Core Flags

```bash
# Log level for Swan: debug, info, warn, error, fatal, panic
# Default: error
SWAN_LOG=error
```

## Kubernetes Flags

These flags control running the experiment workloads on Kubernetes cluster. By default, Swan will run workloads in standalone mode (pure processes).

1. `SWAN_KUBERNETES=true`: Encodes "Kubernetes mode". Swan will launch Kubernetes cluster (kubelet+apiserver+proxy+controller+scheduler) and launch workloads as Kubernetes pods.
1. `SWAN_KUBERNETES_RUN_ON_EXISTING=true`: Runs workloads on cluster provided by user and Swan won't launch it's own cluster. Requires `--kubernetes` flag. Any additional configuration can be provided by `SWAN_KUBERNETES_KUBECONFIG` flag.
1. `SWAN_KUBERNETES_KUBECONFIG`: If launching pods on user-provided cluster requires additional parameters not exposed via flags, user can provide Kubeconfig file. Kubeconfig documentation is provided [here](https://kubernetes.io/docs/concepts/cluster-administration/authenticate-across-clusters-kubeconfig/).


```bash
# Launch HP and BE tasks on Kubernetes.
# Default: false
SWAN_KUBERNETES=false

# Launch HP and BE tasks on existing Kubernetes cluster. (can be use only with --kubernetes flag).
# Default: false
SWAN_KUBERNETES_RUN_ON_EXISTING=false

# Absolute path to the kubeconfig file.
SWAN_KUBERNETES_KUBECONFIG=

# Run experiment pods in selected namespaces.
# Default: default
SWAN_KUBERNETES_NAMESPACE=default

# Run experiment pods on selected node.
# Default: localhost
SWAN_KUBERNETES_NODENAME=localhost

# Comma seperated list of etcd servers in http://ip:port format.
# Default: http://127.0.0.1:2379
SWAN_KUBE_ETCD_SERVERS=http://127.0.0.1:2379
```

## Workloads Flags

### Memcached Flags

1. `SWAN_MEMCACHED_IP`: When the experiment is using external load generators, the user needs to provide address of the interface where Memcached will be listening to.
1. `SWAN_MEMCACHED_THREADS`: Number of Memcached threads. Should be equal to number of full cores provided to Memacached.

```bash
# IP of interface memcached is listening on.
# Default: 127.0.0.1
SWAN_MEMCACHED_IP=127.0.0.1

# Number of threads for mutilate (-t)
# Default: 4
SWAN_MEMCACHED_THREADS=4
```

### Mutilate Flags

1. `SWAN_MUTILATE_MASTER`: Host address where Mutilate master will be launched. Mutilate master is responsible for synchronizing agents and measuring Memcached SLI.

```bash
# Mutilate master host for remote executor. In case of 0 agents being specified it runs in agentless mode.Use `local` to run with local executor.
# Default: 127.0.0.1
SWAN_MUTILATE_MASTER=127.0.0.1

# Mutilate agent threads (-T).
# Default: 8
SWAN_MUTILATE_AGENT_THREADS=8

# Mutilate agent connections (-c).
# Default: 1
SWAN_MUTILATE_AGENT_CONNECTIONS=1

# Mutilate agent hosts for remote executor. Can be specified many times for multiple agents setup.
SWAN_MUTILATE_AGENT=
```

## Experiment Flags

1. `SWAN_AGGR`: Comma separated list of "best effort" workloads that would be launched in colocation with Memcached.
Aggressors available: l1d,l1i,l3d,membw,stream,caffe

```bash
# Aggressor to run experiment with. You can state as many as you want (--aggr=l1d --aggr=membw)
SWAN_AGGR=

# Given SLO for the experiment. [us]
# Default: 500
SWAN_SLO=500

# Number of load points to test
# Default: 10
SWAN_LOAD_POINTS=10

# Load duration [s].
# Default: 10s
SWAN_LOAD_DURATION=10s

# Number of repetitions for each measurement
# Default: 3
SWAN_REPS=3

# Stop experiment in a case of error
# Default: false
SWAN_STOP=false

# Peakload max number of QPS without violating SLO (by default inducted from tuning phase).
# Default: 0
SWAN_PEAK_LOAD=0

```

### Core Isolation Flags

Swan can set core affinity for workloads to show how high priority workload is affected by resource contention.
 
Simple way of declaring cores is through declaring number of cores for each workload.

```bash
# Number of CPUs assigned to high priority task
# Default: 1
SWAN_HP_CPUS=1

# Number of CPUs assigned to best effort task
# Default: 1
SWAN_BE_CPUS=1
```
This way, Swan will run each workload on specific core count. User can expect, that L1 cache aggressors will be run on Memacached workload Hyperthreads, and LLC & Stream aggressors will be launch on different physical cores.


On the other hand, if user wants to handpick cores for workloads, it is possible through those flags:

```bash
# HP cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2
SWAN_HP_SETS=

# BE cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2
SWAN_BE_SETS=

# BE for l1 aggressors cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2
SWAN_BE_L1_SETS=
```
These flags should not be used together with `SWAN_BE_CPUS` and `SWAN_HP_CPUS`.

The last flag is used to show workload interference when real workload is run without any core isolation. Using this flag shows if Linux scheduler or Kubernetes Quality of Service classes are enough for maintaining proper SLO for HP workload.

```bash

# If set, the Caffe workload will use the same isolation settings as for LLC aggressors, otherwise Swan won't apply any performance isolation
# Default: true
SWAN_RUN_CAFFE_WITH_LLCISOLATION=true
```

## Cassandra Flags

These flags contain parameters for connecting to Cassandra DB.

```bash
# Address of Cassandra DB endpoint
# Default: 127.0.0.1
SWAN_CASSANDRA_ADDR=127.0.0.1

# the user name which will be presented when connecting to the cluster
SWAN_CASSANDRA_USERNAME=

# the password which will be presented when connecting to the cluster
SWAN_CASSANDRA_PASSWORD=

# the internal connection timeout for the publisher
# Default: 0s
SWAN_CASSANDRA_TIMEOUT=0s

# determines whether the cassandra publisher should connect to the cluster over an SSL encrypted connection
# Default: false
SWAN_CASSANDRA_SSL=false

# determines whether the publisher will attempt to validate the host
# Default: false
SWAN_CASSANDRA_SSL_HOST_VALIDATION=false

# enables self-signed certificates by setting a certificate authority directly.
SWAN_CASSANDRA_SSL_CA_PATH=

# sets the client certificate, in case the cluster requires client verification
SWAN_CASSANDRA_SSL_CERT_PATH=
# sets the client private key, in case the cluster requires client verification
SWAN_CASSANDRA_SSL_KEY_PATH=
```
