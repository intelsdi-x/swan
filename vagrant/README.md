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

# Local single-node development using Vagrant and Virtualbox

## Quick start

First - install: [Vagrant](https://www.vagrantup.com/docs/installation/), [VirtualBox](https://www.virtualbox.org/wiki/Downloads) and [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git). Then you just need to execute the following commands (they should work on any flavour of Linux):

To run Vagrant development environment, the `$SWAN_DEVELOPMENT_ENVIRONMENT` variable must be set.
If you want to use you local Glide cache in the Vagrant development environment the `$SHARE_GLIDE_CACHE` variable must be set.
```sh
git clone git@github.com:intelsdi-x/swan.git
cd swan/vagrant
export SWAN_DEVELOPMENT_ENVIRONMENT=true
vagrant plugin install vagrant-vbguest  # automatic guest additions
vagrant up
vagrant ssh
```

When you ssh to your virtual machine execute:

```sh
cd swan
make deps build dist install
```

You will need to build [mutilate](https://github.com/leverich/mutilate) by hand and the result binary need to be available in `$PATH`. Consider copying binary to `/opt/swan/bin` and run:
```sh
sudo ln -svf /opt/swan/bin/* /usr/bin/
```

Now you should be able to run an experiment on the Kubernetes cluster or your virtual machine.

If you want to be able to use [iBench](https://github.com/stanford-mast/iBench) they you will need to compile binaries and make then available in `$PATH` (see mutilate description above). Keep in mind that compiling iBench binaries may require a lot of RAM and can't be done on default Vagrant VM configuration.

## Tuning VM parameters

Vagrant will allocate 2 CPUs and 4096 MB RAM for the VM by default. You can consider changing these values but:
* You need to provide at least 2 CPUs.
* You need to provide at least 4096 MB of RAM.

## Deeper dive

If you want to learn more about VM configuration and installed packages refer to [provisioning script](provision.sh). It calls three other scripts and their names should be self explanatory.

Note that the `~/.glide` directory from your host will be mounted on the VM to speed up Go dependency management.

The scripts are responsible for:
* Installing all the necessary CentOS packages that are needed to build Swan, run experiments and analyse their results.
* Installing [Snap](http://snap-telemetry.io/) and its plugins that are responsible for gathering experiment results.
* Installing [Docker](https://www.docker.com/) that allows running experiment on [Kubernetes](https://kubernetes.io) cluster.
* Enabling [Cassandra](http://cassandra.apache.org/) and [Etcd](https://coreos.com/etcd) systemd services.
* Installing [hyperkube](https://github.com/kubernetes/kubernetes/tree/master/cluster/images/hyperkube) that is used to set up Kubernetes cluster for experimantation purposes.
* Setting up SSH for root.
* Installing [Go](https://golang.org/).

If you wish to setup experiment environment on another host then you should be able to run [provision_experiment_environment.sh](provision_experiment_environment.sh) on the host. You will need to provide following environmental variables when calling the script:
* `SWAN_USER` - name of the user that will run experiments.
* `HOME_DIR` - home directory of `SWAN_USER`

Example call can be found in [the provisioning script](provision.sh).

### Provisioners

There are two provisioners defined in the [Vagrantfile](Vagrantfile): `aws` and `virtualbox`. Our CI infracture uses the first of them while the second should be used for development. If you try to use `aws` provider on your own it will fail as AMI is not publicly available.
