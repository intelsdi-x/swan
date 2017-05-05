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

**Docker**

Please install Docker in version 17.03.

```bash
# Installs Docker from docker repository.
# https://docs.docker.com/engine/installation/linux/centos/#install-using-the-repository
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum makecache fast -y -q
sudo yum install -y -q docker-ce-17.03.0.ce-1.el7.centos
sudo echo "Restart docker"
sudo systemctl enable docker
sudo systemctl start docker
```

After installation, please pull Swan image.

```bash
sudo docker pull intelsdi/swan
```

**Experiment Binaries**

Please download Swan binaries from [https://github.com/intelsdi-x/swan/releases](https://github.com/intelsdi-x/swan/releases).
All snap plugins from release package must be in included in `$PATH`.

**Memcached & Best Effort Workloads**

Workloads are deployed from Swan docker image and installed in /opt/swan.

```bash
sudo yum install -y -q glog protobuf boost hdf5 leveldb lmdb opencv libgomp numactl-libs libevent zeromq 
sudo docker run -v /opt:/output intelsdi/swan cp -R /opt/swan /output
```

Path `/opt/swan/bin` must be included in `$PATH`.


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
# Download hyperkube and sets executable bit on it.
curl -O https://storage.googleapis.com/kubernetes-release/release/v1.5.6/bin/linux/amd64/hyperkube
chmod +x hyperkube
cp hyperkube /opt/swan/bin
```

**Snap Plugins**

All plugins must be available in `$PATH`.

```bash
# Downloads Snap plugins and sets executable bit on them. 
sudo wget https://github.com/intelsdi-x/snap-plugin-collector-docker/releases/download/5/snap-plugin-collector-docker_linux_x86_64 -O /opt/swan/bin/snap-plugin-collector-docker
sudo wget https://github.com/intelsdi-x/snap-plugin-collector-use/releases/download/1/snap-plugin-collector-use_linux_x86_64 -O /opt/swan/bin/snap-plugin-collector-use
sudo wget https://github.com/intelsdi-x/snap-plugin-publisher-cassandra/releases/download/5/snap-plugin-publisher-cassandra_linux_x86_64 -O /opt/swan/bin/snap-plugin-publisher-cassandra
sudo wget https://github.com/intelsdi-x/snap-plugin-processor-tag/releases/download/3/snap-plugin-processor-tag_linux_x86_64 -O /opt/swan/bin/snap-plugin-processor-tag
sudo wget https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/2/snap-plugin-publisher-file_linux_x86_64 -O /opt/swan/bin/snap-plugin-publisher-file

sudo chmod +x /opt/swan/bin/snap-plugin-collector-docker
sudo chmod +x /opt/swan/bin/snap-plugin-publisher-cassandra
sudo chmod +x /opt/swan/bin/snap-plugin-processor-tag
sudo chmod +x /opt/swan/bin/snap-plugin-publisher-file
```

**Mutilate**

Mutilate must be compiled from source by user and `mutilate` binary must be available in `$PATH` on Services and Load Generator nodes. Please refer to [Mutilate Readme](https://github.com/leverich/mutilate) for build instructions.

Full list of CentOS dependencies are below. Library cppzmq-devel is required for proper Mutilate agent synchronisation.

```bash
sudo yum install zeromq cppzmq-devel gengetopt libevent-devel scons gcc-c++
# Please clone the https://github.com/leverich/mutilate repository and build it by using `scons`.
# Make sure that cppzmq-devel is installed on all load generator hosts.
```

**Cassandra**

To facilitate Cassandra setup, Swan provides simple systemd service file.
It runs Cassandra docker image and provision it with keyspace and table for Snap metrics ([https://github.com/intelsdi-x/snap-plugin-publisher-cassandra#plugin-database-schema](https://github.com/intelsdi-x/snap-plugin-publisher-cassandra#plugin-database-schema)).

The service file is available [here](https://github.com/intelsdi-x/swan/blob/master/vagrant/cassandra/cassandra.service).

```bash
# Downloads Cassandra service file from Swan repository, adds it to systemd and mounts persistent volume. 
wget https://github.com/intelsdi-x/swan/blob/master/vagrant/cassandra/cassandra.service
sudo mv cassandra.service /etc/systemd/system
sudo mkdir -p /var/data/cassandra
sudo chcon -Rt svirt_sandbox_file_t /var/data/cassandra # SELinux policy
sudo systemctl enable cassandra
sudo systemctl start cassandra
```


For production deployments, please refer to [Datastax Documentation](http://docs.datastax.com/en/landing_page/doc/landing_page/current.html) for details.


**Etcd**

```bash
sudo yum install etcd-3.1.0
```

### Additional Workloads (optional)

Some workloads supported by Swan are not part of default Swan image and their installation is not required for sensitivity profile generation. Workloads binaries should be included in `$PATH` and put inside Swan image (when run on Kubernetes).

**iBench**

[iBench](https://github.com/stanford-mast/iBench) provides synthetic workloads for stressing low level hardware components.

**Stream**

[Stream](https://www.cs.virginia.edu/stream/) is a simple synthetic benchmark program that measures sustainable memory bandwidth (in MB/s) and the corresponding computation rate for simple vector kernels. It can be used to stress memory interconnection.  

## Next
Please move to [Run the Experiment](run_experiment.md) page.
