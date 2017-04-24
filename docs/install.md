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

# ![Swan logo](/images/swan-logo-48.png) Swan

# Installation guide

Swan is built to be run on Linux and has been tested on Linux Centos 7. Instead of building Swan in your own environment, we recommend you to build and run it in a development VM. See [vagrant documentation](../vagrant/README.md) for details.

## Virtual machine configuration details
Swan provides a Vagrantfile, which describes the pre-configured CentOS 7 virtual machine and how to provision it. This machine can be used for running the Swan experiment. The configuration include:

1. Swan directory mounted in the guest file system (it resides in the host OS, but it is also accessed from virtual machine).

2. Additional software packages:
    * docker
    * gengetopt
    * git
    * golang 1.6
    * glide v0.12.2
    * libcgroup-tools
    * libevent-devel
    * nmap-ncat
    * perf
    * scons
    * tree
    * vim
    * wget

    [Glide](https://github.com/Masterminds/glide) is a tool for managing the vendor directory within a Go package. All dependencies are cached in the `~/.glide` folder (this directory is shared between virtual machine and host OS). 

3. A running [docker](https://www.docker.com/) service and a running [Cassandra](http://cassandra.apache.org/) docker container, needed for storing experiment results.

4. [Snap](https://github.com/intelsdi-x/snap) binary placed in `$GOPATH/bin/`.

5. [Kubernetes](http://kubernetes.io/) binaries placed in  `$GOPATH/src/github.com/intelsdi-x/swan/misc/bin`.

6. Workloads binaries placed in `$HOME/swan/workloads/`. To read more about available workloads, please refer to description [here](https://github.com/intelsdi-x/swan/blob/master/experiments/memcached-sensitivity-profile/README.md#aggressor-configuration).

7. Prepared docker swan image `centos_swan_image`, used during experiment, which run on [Kubernetes](http://kubernetes.io/). This image contains all necessary workloads that could be used during experiment (during Vagrant provisioning you will see that workloads are built twice - one time in the virtual machine and the second in the Docker image).

## Prerequisites
You need a read access to [Swan](https://github.com/intelsdi-x/swan) repository. If you don't have it, please contact [Swan](https://github.com/intelsdi-x/swan) repository administrators.

## Install OS dependencies
**1. Git**
The distributed version control system [Git](https://git-scm.com/) is needed to clone the Swan repository. You can install Git as a package, with your local package manager, or download the source code and compile it yourself. See [here](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) for guidance on installation of Git.

**2. VirtualBox**
The cross-platform virtualization application [VirtualBox](https://www.virtualbox.org/) has to be installed because the Swan virtual machine is configured to use this provider. See [here](https://www.virtualbox.org/wiki/Downloads) for guidance on installation of VirtualBox.

**3. Vagrant**
The command line utility for managing the life cycle of virtual machines [Vagrant](https://www.vagrantup.com/docs/) is needed to create and run a pre-configured Swan virtual machine. See [here](https://www.vagrantup.com/docs/installation/) for guidance on installation of Vagrant. *Please note* that Swan requires Vagrant version _1.8.6_.
Configure vagrant to work with VirtualBox by installing the plugin [vagrant-vbguest](https://github.com/dotless-de/vagrant-vbguest) which automatically installs the host's VirtualBox Guest Additions on the guest system:
```
$ vagrant plugin install vagrant-vbguest
$ vagrant box update
```

## Download the Swan sources
```
$ git clone https://github.com/intelsdi-x/swan.git
```

## Prepare SSH configuration
If you do not have SSH keys, generate them. 
Then turn on the ssh-agent and add your keys to the ssh-agent:
```
$ eval "$(ssh-agent -s)"
$ ssh-add ~/.ssh/id_rsa
```
See [here](https://help.github.com/articles/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent/) for further guidance on generating SSH keys and adding them to the ssh-agent.


## Create virtual machine
Create and configure guest machine according to the Swan Vagrantfile. To do this, go to a vagrant directory:
```
$ cd swan/misc/dev/vagrant/singlenode
```
and run:
```
$ vagrant up
```

### Note
It takes ~50 minutes to spin up the whole environment for the first time, so please be patient.
As a result of this command, you will have a running virtual machine pre-configured for running Swan experiment.


## Access Swan code inside virtual machine
To SSH into a running Vagrant machine, go to a vagrant directory:
```bash
$ cd swan/misc/dev/vagrant/singlenode
``` 
and run:
```bash
$ vagrant ssh
```

A Swan repository is placed in the home directory of a vagrant user `$HOME/swan/`

## Build necessary plugins and experiment binary

From within the swan checkout, run the following:
```bash
$ make build_plugins
$ make build_swan
```

This will build and install the [Snap](https://github.com/intelsdi-x/snap) plugins in `$GOPATH/bin/` and the experiment binaries in `$HOME/swan/build/` directory.

## Running tests
1. SSH into the VM: `vagrant ssh`

2. Change to the swan directory: `cd ~/swan`

3. Run the tests: `make test_all`
For further information about tests, please refer to the [Swan development guide](development.md#tests).

## Running experiment
To run the **memcached-sensitivity-profile** experiment, please refer to the [sensitivity experiment README](../experiments/memcached-sensitivity-profile/README.md) for information about how to configure, run it and explore experiment data.

## Changing VM parameters or manually running provision scripts
For details how to change VM parameters or manually run provision scripts, please refer to Vagrant's [README](../misc/dev/vagrant/singlenode/README.md#changing-vm-parameters).

## Troubleshooting
Possible issues that you may encounter are described [here](../misc/dev/vagrant/singlenode/README.md#troubleshooting).
