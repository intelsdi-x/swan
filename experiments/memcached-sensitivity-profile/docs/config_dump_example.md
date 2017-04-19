# Config Dump Example

Sensitivity Profile Experiment has multitude of flags for fine grained experiment control. 
To generate output like this, execute `./memcached-sensitivity-profile -config-dump` command.

Most important flags with description are listed in [Run Experiment](run_experiment.md) page.

```bash
# Export are values.
set -o allexport

# Log level for Swan: debug, info, warn, error, fatal, panic
# Default: error
SWAN_LOG=error

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

# absolute path to the kubeconfig file
SWAN_KUBERNETES_KUBECONFIG=

# run experiment pods in selected namespaces
# Default: default
SWAN_KUBERNETES_NAMESPACE=default

# run experiment pods on selected node
# Default: localhost
SWAN_KUBERNETES_NODENAME=localhost

# Number of lines printed from stderr & stdout in case of task unsucessful termination
# Default: 5
SWAN_OUTPUT_LINES_COUNT=5

# Path to memcached binary
# Default: memcached
SWAN_MEMCACHED_PATH=memcached

# Port for memcached to listen on. (-p)
# Default: 11211
SWAN_MEMCACHED_PORT=11211

# IP of interface memcached is listening on.
# Default: 127.0.0.1
SWAN_MEMCACHED_IP=127.0.0.1

# Username for memcached process (-u)
# Default: root
SWAN_MEMCACHED_USER=root

# Number of threads for mutilate (-t)
# Default: 4
SWAN_MEMCACHED_THREADS=4
                                                  
# Threads affinity (-T) (requires memcached patch)
# Default: false
SWAN_MEMCACHED_THREADS_AFFINITY=false

# Number of maximum connections for mutilate (-c)
# Default: 1024
SWAN_MEMCACHED_CONNECTIONS=1024

# Maximum memory in MB to use for items (-m)
# Default: 64
SWAN_MEMCACHED_MAX_MEMORY=64

# Path to mutilate binary
# Default: mutilate
SWAN_MUTILATE_PATH=mutilate

# Mutilate warmup time [s] (--warmup).
# Default: 10s
SWAN_MUTILATE_WARMUP_TIME=10s

# Mutilate tuning time [s]
# Default: 10s
SWAN_MUTILATE_TUNING_TIME=10s

# Number of memcached records to use (-r).
# Default: 10000
SWAN_MUTILATE_RECORDS=10000

# Mutilate agent threads (-T).
# Default: 8
SWAN_MUTILATE_AGENT_THREADS=8

# Mutilate agent port (-P).
# Default: 5556
SWAN_MUTILATE_AGENT_PORT=5556

# Mutilate agent connections (-c).
# Default: 1
SWAN_MUTILATE_AGENT_CONNECTIONS=1

# Mutilate agent connections (-d).
# Default: 1
SWAN_MUTILATE_AGENT_CONNECTIONS_DEPTH=1

# Mutilate agent affinity (--affinity).
# Default: false
SWAN_MUTILATE_AGENT_AFFINITY=false

# Mutilate agent blocking (--blocking -B).
# Default: true
SWAN_MUTILATE_AGENT_BLOCKING=true

# Mutilate master threads (-T).
# Default: 8
SWAN_MUTILATE_MASTER_THREADS=8

# Mutilate master connections (-C).
# Default: 4
SWAN_MUTILATE_MASTER_CONNECTIONS=4

# Mutilate master connections depth (-C).
# Default: 4
SWAN_MUTILATE_MASTER_CONNECTIONS_DEPTH=4

# Mutilate master affinity (--affinity).
# Default: false
SWAN_MUTILATE_MASTER_AFFINITY=false

# Mutilate master blocking (--blocking -B).
# Default: true
SWAN_MUTILATE_MASTER_BLOCKING=true

# Mutilate master QPS value (-Q).
# Default: 1000
SWAN_MUTILATE_MASTER_QPS=1000

# Length of memcached keys (-K).
# Default: 30
SWAN_MUTILATE_MASTER_KEYSIZE=30

# Length of memcached values (-V).
# Default: 200
SWAN_MUTILATE_MASTER_VALUESIZE=200

# Inter-arrival distribution (-i).
# Default: exponential
SWAN_MUTILATE_MASTER_INTERARRIVALDIST=exponential

# Tail latency Percentile
# Default: 99
SWAN_PERCENTILE=99

# Mutilate master host for remote executor. In case of 0 agents being specified it runs in agentless mode.Use `local` to run with local executor.
# Default: 127.0.0.1
SWAN_MUTILATE_MASTER=127.0.0.1

# Mutilate agent hosts for remote executor. Can be specified many times for multiple agents setup.
SWAN_MUTILATE_AGENT=

# Additional args for kube-apiserver binary (eg. --admission-control="AlwaysAdmit,AddToleration").
SWAN_KUBE_APISERVER_ARGS=

# Additional args for kubelet binary.
SWAN_KUBELET_ARGS=

# Log level for kubernetes servers
# Default: 0
SWAN_KUBE_LOGLEVEL=0

# Allow containers to request privileged mode on cluster and node level (api server and kubelete ).
# Default: false
SWAN_KUBE_ALLOW_PRIVILEGED=false

# Comma seperated list of etcd servers (full URI: http://ip:port)
# Default: http://127.0.0.1:2379
SWAN_KUBE_ETCD_SERVERS=http://127.0.0.1:2379

# Number of checks that kubelet is ready, before trying setup cluster again (with 1s interval between checks).
# Default: 20
SWAN_KUBE_NODE_READY_RETRY_COUNT=20

# Address of a host where Kubernetes master components are to be run
# Default: 127.0.0.1
SWAN_KUBERNETES_MASTER=127.0.0.1

# Address to snapteld in `http://%s:%s` format
# Default: http://127.0.0.1:8181
SWAN_SNAPTELD_ADDRESS=http://127.0.0.1:8181

# Path to script launching caffe as an aggressor
# Default: caffe.sh
SWAN_CAFFE_PATH=caffe.sh

# Path to trained model
# Default: examples/cifar10/cifar10_quick_train_test.prototxt
SWAN_CAFFE_MODEL=examples/cifar10/cifar10_quick_train_test.prototxt

# Path to trained weight
# Default: examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5
SWAN_CAFFE_WEIGHTS=examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5

# Number of iterations
# Default: 1000000000
SWAN_CAFFE_ITERATIONS=1000000000

# Sigint effect for caffe
# Default: stop
SWAN_CAFFE_SIGINT=stop

# Path to L1 Data binary
# Default: l1d
SWAN_L1D_PATH=l1d

# Path to L1 instruction binary
# Default: l1i
SWAN_L1I_PATH=l1i

# Path to L3 Data binary
# Default: l3
SWAN_L3_PATH=l3

# Path to Memory Bandwidth binary
# Default: memBw
SWAN_MEMBW_PATH=memBw

# Number of threads that stream aggressor is going to launch
# Default: 0
SWAN_STREAM_THREAD_NUMBER=0

# Path to stream binary
# Default: stream.100M
SWAN_STREAM_PATH=stream.100M

# set limits and request for HP workloads pods run on kubernetes in CPU millis (default 1000 * number of CPU).
# Default: 8000
SWAN_HP_KUBERNETES_CPU_RESOURCE=8000

# set memory limits and request for HP pods workloads run on kubernetes in bytes (default 1GB).
# Default: 1000000000
SWAN_HP_KUBERNETES_MEMORY_RESOURCE=1000000000

# Launch HP and BE tasks on Kubernetes.
# Default: false
SWAN_KUBERNETES=false

# Launch HP and BE tasks on existing Kubernetes cluster. (can be use only with --kubernetes flag)
# Default: false
SWAN_KUBERNETES_RUN_ON_EXISTING=false

# Aggressor to run experiment with. You can state as many as you want (--aggr=l1d --aggr=membw)
SWAN_AGGR=

# If set, the Caffe workload will use the same isolation settings as for LLC aggressors, otherwise swan won't apply any performance isolation
# Default: true
SWAN_RUN_CAFFE_WITH_LLCISOLATION=true

# Number of L1 data cache aggressors to be run
# Default: 1
SWAN_L1D_PROCESS_NUMBER=1

# Number of L1 instruction cache aggressors to be run
# Default: 1
SWAN_L1I_PROCESS_NUMBER=1

# Number of L3 data cache aggressors to be run
# Default: 1
SWAN_L3_PROCESS_NUMBER=1

# Number of membw aggressors to be run
# Default: 1
SWAN_MEMBW_PROCESS_NUMBER=1

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

# Number of CPUs assigned to high priority task
# Default: 1
SWAN_HP_CPUS=1

# Number of CPUs assigned to best effort task
# Default: 1
SWAN_BE_CPUS=1

# HP cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2
SWAN_HP_SETS=

# BE cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2
SWAN_BE_SETS=

# BE for l1 aggressors cpuset policy with format 'cpuid1,cpuid2:numaid1,numaid2
SWAN_BE_L1_SETS=

# Run baseline phase (without aggressors)
# Default: true
SWAN_BASELINE=true
set +o allexport
                                                  
```
