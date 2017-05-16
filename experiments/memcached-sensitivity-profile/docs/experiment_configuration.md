<!--
 Copyright (c) 2017 Intel Corporation

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# Swan Flags

Most important flags for running Sensitivity Profile are listed here. Full list of flags is shown through `memcached-sensitivity-profile -help` command, or listed in [Example Config Dump](experiment_config_dump_example.md).   

## Swan Core Flags

```bash
# Log level for Swan: debug, info, warn, error, fatal, panic
# Default: info
LOG_LEVEL=info

# Login used for connecting to remote nodes.
# Default value is current user.
# Default: root
REMOTE_SSH_USER=root

# Key for user in from flag "remote_ssh_user" used for connecting to remote nodes.
# Default value is '$HOME/.ssh/id_rsa'
# Default: /root/.ssh/id_rsa
REMOTE_SSH_KEY_PATH=/root/.ssh/id_rsa

# Port used for SSH connection to remote nodes. 
# Default: 22
REMOTE_SSH_PORT=22

```

## Kubernetes Flags

These flags control running the experiment workloads on Kubernetes cluster. By default, Swan will run workloads in standalone mode (pure processes).

1. `KUBERNETES=true`: Encodes "Kubernetes mode". Swan will launch Kubernetes cluster (kubelet+apiserver+proxy+controller+scheduler) and launch workloads as Kubernetes pods.
1. `KUBERNETES_RUN_ON_EXISTING=true`: Runs workloads on cluster provided by user and Swan won't launch it's own cluster. Requires `--kubernetes` flag. Any additional configuration can be provided by `SWAN_KUBERNETES_KUBECONFIG` flag.
1. `KUBERNETES_KUBECONFIG`: If launching pods on user-provided cluster requires additional parameters not exposed via flags, user can provide Kubeconfig file. Kubeconfig documentation is provided [here](https://kubernetes.io/docs/concepts/cluster-administration/authenticate-across-clusters-kubeconfig/).
1. `KUBERNETES_TARGET_NODE_NAME`: When experiment is run on existing Kubernetes cluster, user can point on which node workloads should be launched.


```bash
# Launch Kubernetes cluster and run workloads on Kubernetes. This flag is required to use other kubernetes flags. (caveat: cluster won't be started if `-kubernetes_run_on_existing` flag is set).  
# Default: false
KUBERNETES=false

# Launch HP and BE tasks on existing Kubernetes cluster. (It has to be used with --kubernetes flag). User should provide 'kubernetes_kubeconfig' flag to kubeconfig to point proper API server.
# Default: false
KUBERNETES_RUN_ON_EXISTING=false

# (optional) Absolute path to the kubeconfig file. Overrides pod configuration passed through flags. 
KUBERNETES_KUBECONFIG=

# Experiment's Kubernetes pods will be run on this node. Helpful when experiment is run on existing cluster (KUBERNETES_RUN_ON_EXISTING),
# Default: $HOSTNAME
KUBERNETES_TARGET_NODE_NAME=

# Comma separated list of etcd servers (full URI: http://ip:port)
# Default: http://127.0.0.1:2379
KUBERNETES_CLUSTER_ETCD_SERVERS=http://127.0.0.1:2379

# Kubernetes containers will be run as privileged.
# Default: false
KUBERNETES_PRIVILEGED_PODS=false

# Name of the container image to be used. It needs to be available locally or downloadable.
# Default: intelsdi/swan
KUBERNETES_CONTAINER_IMAGE=intelsdi/swan
```

## Workloads Flags

### Memcached Flags

1. `MEMCACHED_IP`: When the experiment is using external load generators, the user needs to provide address of the interface where Memcached will be listening to.
1. `MEMCACHED_THREADS`: Number of Memcached threads. Should be equal to number of full cores provided to Memacached.
1. `MEMCACHED_THREADS_AFFINITY`: Pins Memcached threads to cores so they are not interfered by scheduler preemption. Memcached supplied by Swan has affinity patch.

```bash
# IP of interface memcached is listening on.
# Default: 127.0.0.1
MEMCACHED_IP=127.0.0.1

# Number of threads for mutilate (-t)
# Default: 4
MEMCACHED_THREADS=4

# Threads affinity (-T) (requires memcached patch)
# Default: false
MEMCACHED_THREADS_AFFINITY=false
```

### Mutilate Flags

1. `SWAN_MUTILATE_MASTER`: Host address where Mutilate master will be launched. Mutilate master is responsible for synchronizing agents and measuring Memcached SLI.
1. `EXPERIMENT_MUTILATE_AGENT_ADDRESSES`: Addresses of machines where Mutilate Load Generators will be launched.

```bash
# Mutilate master host for remote executor. In case of 0 agents being specified it runs in agentless mode.Use `local` to run with local executor.
# Default: 127.0.0.1
EXPERIMENT_MUTILATE_MASTER_ADDRESS=127.0.0.1

# Mutilate agent threads (-T).
# Default: 8
MUTILATE_AGENT_THREADS=8

# Addresses where Mutilate Agents will be launched, separated by commas (e.g: "192.168.1.1,192.168.1.2" Agents generate actual load on Memcached.
# Default: 127.0.0.1
EXPERIMENT_MUTILATE_AGENT_ADDRESSES=192.168.1.1,192.168.1.2

```

### Best Effort Workloads Flags

User can provide his own models for Caffe via `CAFFE_MODEL` and `CAFFE_WEIGHTS` flags.
Also, to pro

```bash
# Path to trained model
# Default: examples/cifar10/cifar10_quick_train_test.prototxt
CAFFE_MODEL=examples/cifar10/cifar10_quick_train_test.prototxt

# Path to trained weights
# Default: examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5
CAFFE_WEIGHTS=examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5

# Number of threads that stream aggressor is going to launch. Default value (0) will launch one thread per cpu.
# Default: 0
EXPERIMENT_BE_STREAM_THREAD_NUMBER=0

# Number of aggressors to be run
# Default: 1
STRESSNG_STREAM_PROCESS_NUMBER=1

# Number of aggressors to be run
# Default: 1
STRESSNG_CACHE_L1_PROCESS_NUMBER=1

# Number of aggressors to be run
# Default: 1
STRESSNG_CACHE_L3_PROCESS_NUMBER=1

# Number of aggressors to be run
# Default: 1
STRESSNG_MEMCPY_PROCESS_NUMBER=1

# Custom arguments to stress-ng
STRESSNG_CUSTOM_ARGUMENTS=

```

## Experiment Flags

1. `EXPERIMENT_BE_WORKLOADS`: Comma separated list of "best effort" workloads that would be launched in colocation with Memcached.

```bash
# Best Effort workloads that will be run sequentially in colocation with High Priority workload. 
# When experiment is run on machine with HyperThreads, user can also add 'stress-ng-cache-l1' to this list. 
# When iBench and Stream is available, user can also add 'l1d,l1i,l3,stream' to this list.
# Default: stress-ng-cache-l3,stress-ng-memcpy,stress-ng-stream,caffe
EXPERIMENT_BE_WORKLOADS=stress-ng-cache-l3,stress-ng-memcpy,stress-ng-stream,caffe

# Given SLO for the HP workload in experiment. [us]
# Default: 500
EXPERIMENT_SLO=500

# Number of load points to test
# Default: 10
EXPERIMENT_LOAD_POINTS=10

# Load duration on HP task.
# Default: 15s
EXPERIMENT_LOAD_DURATION=15s

# Number of repetitions for each measurement
# Default: 3
EXPERIMENT_REPETITIONS=3

# Stop experiment in a case of error
# Default: false
EXPERIMENT_STOP_ON_ERROR=false

# Maximum load that will be generated on HP workload. If value is `0`, then maximum possible load will be found by Swan.
# Default: 0
EXPERIMENT_PEAK_LOAD=0


```

### Core Isolation Flags

Swan can set core affinity for workloads to show how high priority workload is affected by resource contention.
 
Simple way of declaring cores is through declaring number of cores for each workload.

```bash
# Number of CPUs assigned to high priority task. CPUs will be assigned automatically to workloads.
# Default: 1
EXPERIMENT_HP_WORKLOAD_CPU_COUNT=1

# Number of CPUs assigned to best effort task. CPUs will be assigned automatically to workloads.
# Default: 1
EXPERIMENT_BE_WORKLOAD_CPU_COUNT=1
```
This way, Swan will run each workload on specific core count. User can expect, that L1 cache aggressors will be run on Memacached workload Hyperthreads, and LLC & Stream aggressors will be launch on different physical cores.


For debugging purposes, user can set CPU ranges by hand:

```bash
# HP cpuset range (e.g: 0-2). All three 'range' flags must be set to use this policy.
EXPERIMENT_HP_WORKLOAD_CPU_RANGE=

# BE cpuset range (e.g: 0-2) for workloads that are targeted as LLC-interfering workloads. All three 'range' flags must be set to use this policy. 
EXPERIMENT_BE_WORKLOAD_L3_CPU_RANGE=

# BE cpuset range (e.g: 0-2) for workloads that are targeted as L1-interfering workloads. All three 'range' flags must be set to use this policy.
EXPERIMENT_BE_WORKLOAD_L1_CPU_RANGE=
```
These flags should not be used together with `EXPERIMENT_HP_WORKLOAD_CPU_COUNT` and `EXPERIMENT_BE_WORKLOAD_CPU_COUNT`.


The `SWAN_RUN_CAFFE_WITH_LLCISOLATION` flag is used to show workload interference when real workload is run without any core isolation. Using this flag shows if Linux scheduler or Kubernetes Quality of Service classes are enough for maintaining proper SLO for HP workload.

```bash

# If set, the Caffe workload will use the same isolation settings as for LLC aggressors, otherwise Swan won't apply any performance isolation
# Default: true
EXPERIMENT_RUN_CAFFE_WITH_L3_CACHE_ISOLATION=true
```

## Cassandra Flags

These flags contain parameters for connecting to Cassandra DB.

```bash
# Address of Cassandra DB endpoint for Metadata and Snap Publishers.
# Default: 127.0.0.1
CASSANDRA_ADDRESS=127.0.0.1

# The user name which will be presented when connecting to the cluster at 'cassandra_address'.
CASSANDRA_USERNAME=

# The password which will be presented when connecting to the cluster at 'cassandra_address'.
CASSANDRA_PASSWORD=

# Timout for communication with Cassandra cluster.
# Default: 0s
CASSANDRA_TIMEOUT=0s

# Determines whether the cassandra publisher should connect to the cluster over an SSL encrypted connection. Flags CassandraSslHostValidation, CassandraSslCAPath, CassandraSslCertPath and CassandraSslKeyPath should be set accordingly.
# Default: false
CASSANDRA_SSL=false

# Determines whether the publisher will attempt to validate the host. Note that self-signed certificates and hostname mismatch, will cause the connection to fail if not set up correctly. The recommended setting is to enable this flag.
# Default: false
CASSANDRA_SSL_HOST_VALIDATION=false

# Enables self-signed certificates by setting a certificate authority directly. This is not recommended in production settings.
CASSANDRA_SSL_CA_PATH=

# Sets the client certificate, in case the cluster requires client verification.
CASSANDRA_SSL_CERT_PATH=

# Sets the client private key, in case the cluster requires client verification.
CASSANDRA_SSL_KEY_PATH=
```
