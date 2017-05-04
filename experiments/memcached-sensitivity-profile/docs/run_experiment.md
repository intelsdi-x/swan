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


# Running the Experiment

The Sensitivity Profile Experiment can be run in two modes:

1. Standalone - Swan runs workloads as standalone processes.
1. Kubernetes - Swan runs workloads as Kubernetes Pods.

The main difference between these two, is the fact that in Kubernetes mode user can see interferences from Kubernetes and compare them with Standalone mode.


Experiment must be run by privileged user, so that it can set isolation to workloads.
When the experiment is run, an [UUID](https://en.wikipedia.org/wiki/Universally_unique_identifier) like `5df7fa72-add4-44a2-67fa-31668bcafe81` is shown. It will be the identifier for this experiment and the key to retrieve the experiment data.

## Swan Flags

Swan exposes a multitude of flags for fine grained experiment control. To list all flags, plese run `memcached-sensitivity-profile -config-dump`. Dumped config can be later used to run experiment.

```bash
sudo ./memcached-sensitivity-profile -config-dump > config.ini 
sudo ./memcached-sensitivity-profile -config config.ini # Config supplied to experiment.
```

Quick Start configuration is in the next section.

An example config dump is located in [Config Dump Example](config_dump_example.md).
To facilitate experiment instrumentation, the most important Swan are listed in [Swan Flags](swan_flags.md) page.
Flags required for running Experiment in Kubernetes mode are in [Kubernetes Flags](swan_flags.md#Kubernetes-Flags) section.

## Quick Start Configuration

Below is an example configuration using environment variables to set up the experiment where the machines are configured in the following topology:

| Machine       | Role                         |
|:--------------|:-----------------------------|
| 192.168.10.1  | SUT node for Swan Experiment |
| 192.168.10.2  | Load Generator agent node #1 |
| 192.168.10.3  | Load Generator agent node #2 |
| 192.168.10.4  | Load Generator agent node #3 |
| 192.168.10.5  | Load Generator agent node #4 |
| 192.168.10.10 | Services node (also Load Generator Master) |

Binaries should be installed on those machines as stated in [Installation](installation.md) guide.




In this example, SUT node has 32 hyper threads over 16 physical cores on 2 sockets. Per the topology description showed in [Theory](theory.md) section, this leaves 4 threads and logical cores for memcached.
 
Please paste following snippet into `config.ini` file.
Some variables should be changed before running the experiment.

1. `REMOTE_SSH_LOGIN` and `REMOTE_SSH_KEY_PATH` should point to user and key that is authorized to SSH on every machine in the experiment cluster.
1. `MEMCACHED_USER` should be the same as `REMOTE_SSH_LOGIN`
1. `MEMCACHED_LISTENING_ADDRESS` should be SUT node address.
1. `EXPERIMENT_PEAK_LOAD` should be provided appropriate to SUT machine. If tiny VM is used, it might be even as low as 100000. If high-end server machine is used, it might be as high as millions requests per second.
1. `EXPERIMENT_SLO` should be set appropriate to SUT machine (in microseconds). In VM environment, the default value should be increased 10x. 
1. When SUT node has HyperThreads, `EXPERIMENT_BE_WORKLOADS` should contain `stress-ng-cache-l1` entry to examine Memcached sensitivity to L1 cache interference. 
 
 

```ini
# Log level for Swan: debug, info, warn, error, fatal, panic
# Default: info
LOG_LEVEL=info

## --- Remote SSH Access ---
REMOTE_SSH_USER=root
REMOTE_SSH_KEY_PATH=/root/.ssh/id_rsa

## --- Best Effort Workloads ---
# Best Effort workloads that will be run sequentially in colocation with High Priority workload. 
# When experiment is run on machine with HyperThreads, user can also add 'stress-ng-cache-l1' to this list. 
EXPERIMENT_BE_WORKLOADS=stress-ng-cache-l3,stress-ng-memcpy,stress-ng-stream,caffe
EXPERIMENT_RUN_CAFFE_WITH_L3_CACHE_ISOLATION=false

## --- Experiment configuration ---
# Highiest load that SUT machine can handle without breaking the SLO.
EXPERIMENT_PEAK_LOAD=600000
# Given SLO for the HP workload in experiment in microseconds.
EXPERIMENT_SLO=500
# Duration of each measurement.
EXPERIMENT_LOAD_DURATION=15s
# Each load point is fraction of peak load.
EXPERIMENT_LOAD_POINTS=10
# Number of times each load point will be repeated.
EXPERIMENT_REPETITIONS=1

## --- Isolation ---
EXPERIMENT_HP_WORKLOAD_CPU_COUNT=4
EXPERIMENT_BE_WORKLOAD_CPU_COUNT=4

## --- Memcached Configuration ---
MEMCACHED_LISTENING_ADDRESS=192.168.10.1
MEMCACHED_USER=root
MEMCACHED_THREADS=4
MEMCACHED_THREADS_AFFINITY=false

## --- Mutilate Configuration ---
# Master
EXPERIMENT_MUTILATE_MASTER_ADDRESS=192.168.10.10
MUTILATE_MASTER_THREADS=8
MUTILATE_MASTER_CONNECTIONS=8

# Agents
EXPERIMENT_MUTILATE_AGENT_ADDRESSES=192.168.10.2,192.168.10.3,192.168.10.4,192.168.10.5

## --- Snap Configuration ---
SNAPTELD_ADDRESS=http://192.168.10.1:8181

## --- Cassandra Configuration ---
CASSANDRA_ADDRESS=192.168.10.10

## --- Kubernetes Configuration ---
# Uncomment following flags to run workloads on Kubernetes.
# Experiment will ramp up cluster on SUT and Services node. 
# KUBERNETES=true
# KUBERNETES_CLUSTER_RUN_CONTROL_PLANE_ON_HOST=192.168.10.10
```

Before running `memcached-sensitivity-profile` please ensure that
* Cassandra is up and running on the Services node.
* Snapteld is running on SUT.
* Mutilate binary and `cppzmq-devel` library are installed on Service and Load Generator hosts.
* From SUT node, an user passed in `REMOTE_SSH_USER` flag can connect via ssh to other nodes using keys authorization.

If everything is ready then simply launch:

```
sudo ./memcached-sensitivity-profile -config config.ini
```

Note the UUID that is printed on stdout and wait for experiment to finish.

## Explore Experiment Data (Sensitivity Profile)

When the experiment is complete, the results can be retrieved from Cassandra.
Swan ships with a Jupyter Notbook which provides an environment for loading the samples and generating sensitivity profiles.
For instructions on how to run Jupyter Notebook, please refer to the [Jupyter user guide](../../../jupyter/README.md).

A few pointers to validate the experiment data:

 - Baseline measurements should not violate SLO at any load point.
 - At low loads - numbers may not differ for baseline and colocated scenarios. The differences should be in _when_ the saturation occurs. For the colocated scenarios, this should become evident at higher loads. If this does not occur, it might mean that Memcached has not been properly baselined.

Below is an example of what the sensitivity profile could be:

![Sensitivity profile](/images/sensitivity-profile.png)


The _Load_ row is a percentage of the peak load which was found during _[Red lining](https://www.wikiwand.com/en/Redline)_.
A cell in a table express _SLI_ which is a 99th [percentile](https://www.wikiwand.com/en/Percentile) response time for that _Load_ in relation to _SLO_. For instance _Baseline_ for _Load_ 5% for _SLO_ 500ms tells that 99 percent of requests responded in time not greater than 160ms which is 32% of SLO time. Thus if we observe _SLI_ above 100% that means violation of _SLO_.
In the presented table Caffe and memBW are relatively weak aggressors and they lead to _SLO_ violation only on higher loads while Stream 100M is very aggressive and leads to _SLO_ violation even on low loads of memcached.

## Next
Please move to [Tuning](tuning.md) page.

To see all available flags, please look at [Config Dump Example](config_dump_example.md).
