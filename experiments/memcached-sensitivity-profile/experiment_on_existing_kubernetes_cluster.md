# ![Swan logo](../../docs/swan-logo-48.png) Swan 

# Memcached Sensitivity Profile on Existing Kubernetes Cluster

Memcached Sensitivity Profile can be run on your own Kubernetes cluster. This instruction will walk you through preparation of machines and necesary components.

## Preparation
To make sure that results from this experiment will be sound, you should split your infrastructure on three parts:
1. System Under Test (SUT) - machine where Memcached will be collocated with aggressors. 
1. Load Generator Machines - machines which generates load on Memcached.
1. Load Generator Master - machine for SLI measurmenet of Memcached.
1. Utilities - database node for storing experiment results.

You can run experiment with all components on single host but be aware that sensitivity profile might be skewed by control plane and load generation interference. Also you won't see latencies added by network hardware. If this experiment is run on virtual machine, then user could not access specific hardware metrics eg. last level cache occupancy and memory bandwidth usage.

Default database for this experiment is Cassandra in version 3.x

### Hardware Requirements
You should have 10G network between Load Generators and Memcached host to saturate Memacached.

### Software Requirements

This experiment is tested and supported on CentOS 7.2. Experiment requires SSH access to `root` user from SUT to all other machines used in experiment. 

## Software Used
Services required by Experiment:
1. Docker (version supported by user's Kubernetes) 
1. Etcd (version supported by user's Kubernetes)
1. Kubernetes in version 1.5.x
1. Cassandra in version 3.x
1. Snap in version 1.1

Binaries run by Experiment:
1. Mutilate

Dockerfile run by Experiment should contain these applications:
1. Memcached
1. iBench
1. Caffe

Snap plugins used by Experiment:
1. TBD

### Software preparations
This experiments expects that Kubernetes cluster with Docker and etcd are already in place.
#### Services Installation
##### Cassandra
Please follow instructions on http://cassandra.apache.org/doc/latest/ to install Cassandra on your cluster. You can also run it inside Docker container - instructions can be found here: https://hub.docker.com/_/cassandra/.

When cassandra is ready, please create Snap keyspace: 
```
CREATE KEYSPACE IF NOT EXISTS snap WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
DESCRIBE KEYSPACES;
```
And prepare metric table as described here: https://github.com/intelsdi-x/snap-plugin-publisher-cassandra#plugin-database-schema

##### Snap
Please follow instructions listed here https://github.com/intelsdi-x/snap#installation to install Snap on SUT and Utilities node.

#### Binaries Installation
Please make sure that all binaries listed here are available in path for user `root`.

##### Mutilate
Please follow instructions on https://github.com/leverich/mutilate to build Mutilate and make sure that it is available in path on all nodes used in load generation.
Please make sure that it is compiled with LibZMQ support.

##### Snap Plugins
TBD.

#### Docker Image Preparation
User should create Docker Image with all application listed here available in path. Image should be named`swan-docker-image` and should be based on Centos7 image. 

##### Memcached
Please follow instructions on http://memcached.org/downloads to install Memcached.

##### iBench
Please download all files from https://github.com/stanford-mast/iBench and run `make` to build the binaries. Make sure they will be available in Path.

##### Caffe
Please follow instructions on http://caffe.berkeleyvision.org/installation.html to install Caffe. Plase install it in `/opt/swan/share/caffe` directory.
Make sure that:
1. Caffe is compiled on CPU_ONLY mode
1. Caffe is compiled with OpenBLAS

After installation, please prepare CIFAR10 dataset as instructed here: http://caffe.berkeleyvision.org/gathered/examples/cifar10.html.
Please set `solver_mode` to CPU as instructed in http://caffe.berkeleyvision.org/gathered/examples/cifar10.html.

To finish installation, please add `caffe.sh` script in path.

```bash
#!/bin/bash
set -e

CAFFE_DIR=/opt/swan/share/caffe

if [ ! -x ${CAFFE_DIR}/bin/caffe ] ; then
    echo "error: caffe has to be installed $CAFFE_DIR first!"
    exit 1
fi

cd $CAFFE_DIR
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$CAFFE_DIR/lib
./bin/caffe "$@"
```

### Running the experiment
Please read main experiment README.

Before running experiment please fill kubeconfig file with all necessary parameters. Kubeconfig file documentation is located here: https://kubernetes.io/docs/user-guide/kubectl/kubectl_config/ 

Then, please run the experiment with these flags:
```shell
--kubernetes=true
--run-on-kubernetes=true
--kubernetes-nodename=<node_where_workloads_will_be_colocated>
--kubernetes-config=<path_to_kubeconfig_file> 
```
