# Installation guide for Swan
Swan is built to be run on Linux and has been tested on Linux Centos 7.
Instead of building Swan in your own environment, we recommend you to build and run it in a development VM. Follow the instructions to create a Linux virtual machine pre-configured for running the Swan experiment.

## Prerequities
You have to have a read access to [Athena](https://github.com/intelsdi-x/athena) and [Swan](https://github.com/intelsdi-x/swan) repositories. If you don't have it, please contact [Swan](https://github.com/intelsdi-x/swan) repository administrators.

## Install OS dependencies
**1. Git**
The distributed version control system [Git](https://git-scm.com/) is needed to clone Swan repository. You can install Git as a package, via another installer, or download the source code and compile it yourself. See [here](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) for guidance on installation of Git.

**2. VirtualBox**
The cross-platform virtualization application [VirtualBox](https://www.virtualbox.org/) has to be installed because the Swan virtual machine is configured to use this provider. See [here](https://www.virtualbox.org/wiki/Downloads) for guidance on installation of VirtualBox.

**3. Vagrant**
The command line utility for managing the lifecycle of virtual machines [Vagrant](https://www.vagrantup.com/docs/) is needed to create and run a pre-configured Swan virtual macine. See [here](https://www.vagrantup.com/docs/installation/) for guidance on installation of Vagrant.
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

Make sure your public SSH key is added to your github account, if not follow the instructions [here](https://help.github.com/articles/adding-a-new-ssh-key-to-your-github-account/) to add it.

## Create virtual machine
Create and configure guest machine according to the Swan Vagrantfile. To do this, go to a vagrant directory:
```
$ cd swan/misc/dev/vagrant/singlenode
```
and run:
```
$ vagrant up
```
As a result of this command, you will have a running virtual machine pre-configured for running Swan experiment.
Configuration include:

1. A running [Cassandra](http://cassandra.apache.org/) docker, needed for storing experiment results.

2. [Snap](https://github.com/intelsdi-x/snap) binary placed in `$GOPATH/bin/`.

3. Workloads binaries placed in `$HOME/swan/workloads/`. To read more about available workloads, please refer to description [here](https://github.com/intelsdi-x/swan/blob/master/experiments/memcached-sensitivity-profile/README.md#aggressor-configuration).

## Access Swan code inside virtual machine
To SSH into a running Vagrant machine, run:
```
$ vagrant ssh
```
A Swan repository is placed in the home directory of a vagrant user `$HOME/swan/`

## Build necessary plugins and experiment binary

From within the swan checkout, run the following:
```
$ make build_plugins
$ make build_swan
```

This will build and install the [Snap](https://github.com/intelsdi-x/snap) plugins in `$GOPATH/bin/` and the experiment binaries in `$HOME/swan/build/` directory.

To run tests, please refer to the [Swan development guide](development.md).

To run the **memcached-sensitivity-profile** experiment, please refer [here](experiments/memcached-sensitivity-profile/README.md) for information about how to configure, run it and explore experiment data.

