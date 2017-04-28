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

# Swan Experiment Installation
This guide will walk you through installation of required binaries and services to successfully run Sensitivity Profile Experiment.

As written in [Prerequisites](prerequisites.md) section, we have three classes of nodes:
 * System Under Test (SUT)
 * Load Generator Agents
 * Services Node
 
Single node can have multiple classes, but this separation is recommended to obtain correct results from experiment.

### Required Software

**System Under Test Node:**

SUT Node requires following application installed.

* Experiment Binaries - *Swan Experiments*
* Memcached & Best Effort Workloads - *Worklods that will be tested in colocation*
* Snap - *Intel Telemetry Framework*
* Hyperkube - *Kubernetes in single binary*
* Docker - *Container Runtime*
* Snap Plugins - *Plugins for gathering telemetry & experiment results*

**Load Generator Agent:**

* Mutilate - *Memcached Load Generator*

**Services Node:**

* Mutilate - *Master for agent synchronisation*
* Cassandra - *Database for storing experiment results* 
* Hyperkube - *Kubernetes Control Plane*
* Etcd - *Backend for Kubernetes*

### The Installation details are as follow:

`wget` must be installed on node.

**Experiment Binaries**

Please download Swan binaries from [https://github.com/intelsdi-x/swan/releases](https://github.com/intelsdi-x/swan/releases).

**Memcached & Best Effort Workloads**

Workloads are deployed from Swan docker image and installed in /opt/swan.
Deployed binaries must be available in `$PATH`.

```bash
sudo docker run -v `pwd`/opt:/output intelsdi/swan cp -R /opt/swan /output
```

Following workloads are installed this way:
* Memcached 1.4.35 with *thread affinity* patch
* Stress-ng - *Synthethic stresser that can stress system in various selectable ways*
* Caffe - *Deep Learning framework to simulate real workload*

Swan also supports [iBench](https://github.com/stanford-mast/iBench) and [Stream Benchmark](https://www.cs.virginia.edu/stream/) workloads that are not deployed by preceding script. Stress-ng can stress system in similar way. 

**Snap**

Snap installation instruction are available [here](https://github.com/intelsdi-x/snap#installation).


```bash
curl -s https://packagecloud.io/install/repositories/intelsdi-x/snap/script.rpm.sh | sudo bash
sudo yum install -y snap-telemetry
sudo systemctl enable snap-telemetry.service
sudo systemctl start snap-telemetry.service
```

**Hyperkube**

Please download Hyperkube binary and put it in `$PATH` on SUT and Service nodes.

```bash
curl -O https://storage.googleapis.com/kubernetes-release/release/v1.5.6/bin/linux/amd64/hyperkube
chmod +x hyperkube
```

**Docker**

Please install Docker in version 17.03.

```bash
# https://docs.docker.com/engine/installation/linux/centos/#install-using-the-repository
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum makecache fast -y -q
sudo yum install -y -q docker-ce-17.03.0.ce-1.el7.centos
sudo echo "Restart docker"
sudo systemctl enable docker
sudo systemctl start docker
```

**Snap Plugins**

All plugins must be available in `$PATH`.

```bash
wget https://github.com/intelsdi-x/snap-plugin-collector-docker/releases/download/5/snap-plugin-collector-docker_linux_x86_64 -O snap-plugin-collector-docker
wget https://github.com/intelsdi-x/snap-plugin-collector-use/releases/download/1/snap-plugin-collector-use_linux_x86_64 -O snap-plugin-collector-use
wget https://github.com/intelsdi-x/snap-plugin-publisher-cassandra/releases/download/5/snap-plugin-publisher-cassandra_linux_x86_64 -O snap-plugin-publisher-cassandra
wget https://github.com/intelsdi-x/snap-plugin-processor-tag/releases/download/3/snap-plugin-processor-tag_linux_x86_64 -O snap-plugin-processor-tag
wget https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/2/snap-plugin-publisher-file_linux_x86_64 -O snap-plugin-publisher-file

chmod +x snap-plugin-collector-docker
chmod +x snap-plugin-publisher-cassandra
chmod +x snap-plugin-processor-tag
chmod +x snap-plugin-publisher-file
```

**Mutilate**

Mutilate must be compiled from source by user and `mutilate` binary must be available in `$PATH` on Services and Load Generator nodes. Please refer to [Mutilate Readme](https://github.com/leverich/mutilate) for build instructions.

Full list of CentOS dependencies are below. Library cppzmq-devel is required for proper Mutilate agent synchronisation.

```bash
sudo yum install cppzmq-devel gengetopt libevent-devel scons gcc-c++
# Plese clone the https://github.com/leverich/mutilate repository and build it by using `scons`
```

**Cassandra**

For example experiments, simple Dockerized Cassandra should be enough. Please see [Simple Cassandra Installation](simple_cassandra_installation.md.go) for details.

For production deployments, please refer to [Datastax Documentation](http://docs.datastax.com/en/landing_page/doc/landing_page/current.html) for details.


**Etcd**

```bash
sudo yum install etcd-3.1.0
```

## Next
Please move to [Run the Experiment](run_experiment.md) page.
