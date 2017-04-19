# Swan Experiment Installation
This guide will walk you through installation of required binaries and services to successfully run Sensitivity Profile Experiment.

## Service Installation
These services are required for Experiment operation. 

### Cassandra Service
Please install Cassandra on the Service node.

Swan is compatible with Cassandra version 3.x and requires keyspaces `swan` and `snap` to be available.

If Cassandra is not already available please follow instructions on [Datastax's installation guide](http://docs.datastax.com/en/cassandra/3.x/cassandra/cassandraAbout.html) or run it on Docker using official image: https://hub.docker.com/_/cassandra/

**NOTE** Running Cassandra in docker containers is not advised for production environments.
Additionally, be careful if not used with docker volume mounts as you may experience data loss.

In order to reduce any unintended interference, it is not recommended that any of the machines involved in the experiment (executing memcached or load generators) should also be host for Cassandra. 

### Snap Service
Please Install Snap on System Under Test (SUT) machine.

Swan uses Snap to collect and store metrics in Cassandra Database. See the [Snap installation guide](https://github.com/intelsdi-x/snap#installation) for guidance of how to configure and install Snap. 

## Binaries Installation
All of binaries listed here should be available in directory listed in `$PATH` for user `root`. They should be deployed on SUT, except Mutilate which should be deployed on all Load Generator and Services machines.

### Kubernetes (optional)
When Experiment is not planned to be run on Kubernetes, or when user provides it's own cluster this step can be skipped. Instruction containing connection to user provided Kubernetes cluster are described in [Kubernetes Flags](swan_flags.md#Kubernetes-Flags) section.

The Experiment requires Kubernetes in version 1.5.x. 
Please download binaries from https://github.com/kubernetes/kubernetes/releases, copy hyperkube to directory listed in`$PATH` and run `./hyperkube --make-symlinks` to make it usable by Swan. Hyperkube binary should have executable bit set. Swan expects that Docker will be the default container runtime for Kubernetes. Please make sure that all dependencies for Kubernetes (e.g. Docker, Etcd) are running and in proper versions.

Binaries should be placed in the directory listed in `$PATH` for `root` user on System Under Test (SUT) and Service machines.
 
#### Docker Image (only when experiments are going to be run on Kubernetes)
User should create the Docker Image with all applications listed below available in `$PATH`. Image should be named `centos-swan-image` and should be based on Centos7 image. It should be available in local Docker images on SUT machine.

### Memcached
Please follow instructions on http://memcached.org/downloads to install Memcached.

It should be available for user `root` in directory listed in `$PATH` on SUT node. 

### iBench
Please download all files from https://github.com/stanford-mast/iBench and run `make` to build the binaries. Make sure binaries will be available in directory listed in `$PATH`.

It should be available for user `root` in directory listed in `$PATH` on SUT node.

### Caffe
Please follow instructions on http://caffe.berkeleyvision.org/installation.html to install Caffe. Plase install it in `/opt/swan/share/caffe` directory.
Make sure that:
1. Caffe is compiled on CPU_ONLY mode
1. Caffe is compiled with OpenBLAS

After installation, please prepare CIFAR10 dataset as instructed here: http://caffe.berkeleyvision.org/gathered/examples/cifar10.html.
Please set `solver_mode` to CPU as instructed in http://caffe.berkeleyvision.org/gathered/examples/cifar10.html.

To finish installation, please add `caffe.sh` script in directory listed in `$PATH` on SUT machine:

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

### Mutilate
Please follow instructions on https://github.com/leverich/mutilate to install Mutilate.

It should be available for user `root` in directory listed in `$PATH` on Load Generator machines.

## Experiment Binaries

Please download binaries from [Releases](https://github.com/intelsdi-x/swan/releases) page on Github. Snap plugins from package should be added to directory listed in `$PATH`.

## Next
Please move to [Run the Experiment](run_experiment.md) page.
