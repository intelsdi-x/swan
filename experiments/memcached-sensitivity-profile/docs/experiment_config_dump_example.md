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

# Config Dump Example

Sensitivity Profile Experiment has multitude of flags for fine grained experiment control. 
To generate output like this, execute `./memcached-sensitivity-profile -config-dump` command.

Most important flags with description are listed in [Run Experiment](run_experiment.md) page.

```bash

# Log level for Swan: debug, info, warn, error, fatal, panic
# Default: info
LOG_LEVEL=info

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

# (optional) Absolute path to the kubeconfig file. Overrides pod configuration passed through flags. 
KUBERNETES_KUBECONFIG=

# Kubernetes containers will be run as privileged.
# Default: false
KUBERNETES_PRIVILEGED_PODS=false

# Kubernetes Pod launch timeout.
# Default: 30s
KUBERNETES_POD_LAUNCH_TIMEOUT=30s

# Name of the container image to be used. It needs to be available locally or downloadable.
# Default: intelsdi/swan
KUBERNETES_CONTAINER_IMAGE=intelsdi/swan

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

# Port for Memcached to listen on. (-p)
# Default: 11211
MEMCACHED_PORT=11211

# IP address of interface that Memcached will be listening on. It must be actual device address, not '0.0.0.0'.
# Default: 127.0.0.1
MEMCACHED_LISTENING_ADDRESS=127.0.0.1

# Username for Memcached process. (-u)
# Default: root
MEMCACHED_USER=root

# Number of threads to use. (-t)
# Default: 4
MEMCACHED_THREADS=4

# Threads affinity (-T) (requires memcached patch)
# Default: false
MEMCACHED_THREADS_AFFINITY=false

# Max simultaneous connections. (-c)
# Default: 2048
MEMCACHED_CONNECTIONS=2048

# Maximum memory in MB to use for items in megabytes. (-m)
# Default: 4096
MEMCACHED_MAX_MEMORY=4096

# Mutilate tuning time [s].
# Default: 10s
MUTILATE_TUNING_TIME=10s

# Mutilate warmup time [s] (--warmup).
# Default: 1s
MUTILATE_WARMUP_TIME=1s

# Number of memcached records to use (-r).
# Default: 5000000
MUTILATE_RECORDS=5000000

# Mutilate agent threads (-T).
# Default: 8
MUTILATE_AGENT_THREADS=8

# Mutilate agent port (-P).
# Default: 5556
MUTILATE_AGENT_PORT=5556

# Mutilate agent connections (-c).
# Default: 16
MUTILATE_AGENT_CONNECTIONS=16

# Mutilate agent connections (-d).
# Default: 1
MUTILATE_AGENT_CONNECTIONS_DEPTH=1

# Mutilate agent affinity (--affinity).
# Default: true
MUTILATE_AGENT_AFFINITY=true

# Mutilate agent blocking (--blocking -B).
# Default: true
MUTILATE_AGENT_BLOCKING=true

# Mutilate master threads (-T).
# Default: 8
MUTILATE_MASTER_THREADS=8

# Mutilate master connections (-C).
# Default: 4
MUTILATE_MASTER_CONNECTIONS=4

# Mutilate master connections depth (-D).
# Default: 1
MUTILATE_MASTER_CONNECTIONS_DEPTH=1

# Mutilate master affinity (--affinity).
# Default: true
MUTILATE_MASTER_AFFINITY=true

# Mutilate master blocking (--blocking -B).
# Default: true
MUTILATE_MASTER_BLOCKING=true

# Mutilate master QPS value (-Q).
# Default: 1000
MUTILATE_MASTER_QPS=1000

# Length of memcached keys (-K).
# Default: 30
MUTILATE_MASTER_KEYSIZE=30

# Length of memcached values (-V).
# Default: 200
MUTILATE_MASTER_VALUESIZE=200

# Inter-arrival distribution (-i).
# Default: exponential
MUTILATE_MASTER_INTERARRIVAL_DIST=exponential

# Tail latency percentile for Memcached SLI
# Default: 99
EXPERIMENT_TAIL_LATENCY_PERCENTILE=99

# Address where Mutilate Master will be launched. Master coordinate agents and measures SLI.
# Default: 127.0.0.1
EXPERIMENT_MUTILATE_MASTER_ADDRESS=127.0.0.1

# Addresses where Mutilate Agents will be launched, separated by commas (e.g: "192.168.1.1,192.168.1.2" Agents generate actual load on Memcached.
# Default: 127.0.0.1
EXPERIMENT_MUTILATE_AGENT_ADDRESSES=127.0.0.1

# Comma seperated list of etcd servers (full URI: http://ip:port)
# Default: http://127.0.0.1:2379
KUBERNETES_CLUSTER_ETCD_SERVERS=http://127.0.0.1:2379

# Address of a host where Kubernetes control plane will be run (when using -kubernetes and not connecting to existing cluster).
# Default: 127.0.0.1
KUBERNETES_CLUSTER_RUN_CONTROL_PLANE_ON_HOST=127.0.0.1

# Snapteld address in `http://%s:%s` format
# Default: http://127.0.0.1:8181
SNAPTELD_ADDRESS=http://127.0.0.1:8181

# Path to trained model
# Default: examples/cifar10/cifar10_quick_train_test.prototxt
CAFFE_MODEL=examples/cifar10/cifar10_quick_train_test.prototxt

# Path to trained weights
# Default: examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5
CAFFE_WEIGHTS=examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5

# Number of threads that stream aggressor is going to launch. Default value (0) will launch one thread per cpu.
# Default: 0
EXPERIMENT_BE_STREAM_THREAD_NUMBER=0

# Custom arguments to stress-ng
STRESSNG_CUSTOM_ARGUMENTS=

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

# Launch Kubernetes cluster and run workloads on Kubernetes. This flag is required to use other kubernetes flags. (caveat: cluster won't be started if `-kubernetes_run_on_existing` flag is set).  
# Default: false
KUBERNETES=false

# Launch HP and BE tasks on existing Kubernetes cluster. (It has to be used with --kubernetes flag). User should provide 'kubernetes_kubeconfig' flag to kubeconfig to point proper API server.
# Default: false
KUBERNETES_RUN_ON_EXISTING=false

# Sets CPU resource limit and request for HP workload on Kubernetes [CPU millis, default 1000 * number of CPU].
# Default: 8000
KUBERNETES_HP_CPU_RESOURCE=8000

# Sets memory limit and request for HP workloads on Kubernetes in bytes (default 1GB).
# Default: 1000000000
KUBERNETES_HP_MEMORY_RESOURCE=1000000000

# Experiment's Kubernetes pods will be run on this node.
# Default: ubuntus
KUBERNETES_TARGET_NODE_NAME=ubuntus

# Best Effort workloads that will be run sequentially in colocation with High Priority workload. 
# When experiment is run on machine with HyperThreads, user can also add 'stress-ng-cache-l1' to this list. 
# When iBench and Stream is available, user can also add 'l1d,l1i,l3,stream' to this list.
# Default: stress-ng-cache-l3,stress-ng-memcpy,stress-ng-stream,caffe
EXPERIMENT_BE_WORKLOADS=stress-ng-cache-l3,stress-ng-memcpy,stress-ng-stream,caffe

# Debug only: Best Effort workloads are wrapped in Service flags so that the experiment can track their lifectcle. Default `true` should not be changed without explicit reason.
# Default: true
DEBUG_TREAT_BE_AS_SERVICE=true

# Number of L1 data cache best effort processes to be run
# Default: 1
EXPERIMENT_BE_L1D_PROCESSES_NUMBER=1

# Number of L1 instruction cache best effort processes to be run
# Default: 1
EXPERIMENT_BE_L1I_PROCESSES_NUMBER=1

# Number of L3 data cache best effort processes to be run
# Default: 1
EXPERIMENT_BE_L3_PROCESSES_NUMBER=1

# Number of membw best effort processes to be run
# Default: 1
EXPERIMENT_BE_MEMBW_PROCESSES_NUMBER=1

# If set, the Caffe Best Effort workload will use the same isolation settings as for L3 Best Efforts, otherwise swan won't apply any performance isolation. User can use this flag to compare running task on separate cores and using OS scheduler.
# Default: true
EXPERIMENT_RUN_CAFFE_WITH_L3_CACHE_ISOLATION=true

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

# Amount of time to wait for load generator to stop before stopping it forcefully. In succesful case, it should stop on it's own.
# Default: 0s
EXPERIMENT_LOAD_GENERATOR_WAIT_TIMEOUT=0s

# Number of CPUs assigned to high priority task. CPUs will be assigned automatically to workloads.
# Default: 1
EXPERIMENT_HP_WORKLOAD_CPU_COUNT=1

# Number of CPUs assigned to best effort task. CPUs will be assigned automatically to workloads.
# Default: 1
EXPERIMENT_BE_WORKLOAD_CPU_COUNT=1

# HP cpuset range (e.g: 0-2). All three 'range' flags must be set to use this policy.
EXPERIMENT_HP_WORKLOAD_CPU_RANGE=

# BE cpuset range (e.g: 0-2) for workloads that are targeted as LLC-interfering workloads. All three 'range' flags must be set to use this policy. 
EXPERIMENT_BE_WORKLOAD_L3_CPU_RANGE=

# BE cpuset range (e.g: 0-2) for workloads that are targeted as L1-interfering workloads. All three 'range' flags must be set to use this policy.
EXPERIMENT_BE_WORKLOAD_L1_CPU_RANGE=
```
