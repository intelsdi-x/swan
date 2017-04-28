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

```sh
git clone git@github.com:intelsdi-x/swan.git
cd swan/vagrant
vagrant plugin install vagrant-vbguest  # automatic guest additions
vagrant box update
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

Now you should be able to run an experiment on the Kubernetes cluster that was automatically provisioned. If you want to be able to run without Kuberenetes, then you will need to take a few more steps.

Some of the project dependencies are distributed as [Docker image](https://hub.docker.com/r/intelsdi/swan/). They need to be extracted from the container in order to be used for non-Kubernetes experiments. To do this you need to execute the following commands:

```sh
make extract_binaries
sudo chown -R $USER:$USER opt 
sudo cp -fa opt/swan /opt
sudo ln -svf /opt/swan/bin/* /usr/bin/
```

If you want to be able to use [iBench](https://github.com/stanford-mast/iBench) they you will need to compile binaries and make then available in `$PATH` (see mutilate description above). Keep in mind that compiling iBench binaries may require a lot of RAM and can't be done on default Vagrant VM configuration.

## Tuning VM parameters

Vagrant will allocate 2 CPUs and 4096 MB RAM for the VM by default. You can consider changing these values but:
* You need to provide at least 2 CPUs.
* You need to provide at least 4096 MB of RAM.

## Deeper dive

If you want to learn more about VM configuration and installed packages refer to [provisioning script](provision.sh).

Note that the `~/.glide` directory from your host will be mounted on the VM to speed up Go dependency management.

The script is responsible for:
* Installing all the necessary CentOS packages that are needed to build Swan, run experiments and analyse their results.
* Installing [Snap](http://snap-telemetry.io/) and its plugins that are responsible for gathering experiment results.
* Installing [Docker](https://www.docker.com/) that allows running experiment on [Kubernetes](https://kubernetes.io) cluster.
* Enabling [Cassandra](http://cassandra.apache.org/) and [Etcd](https://coreos.com/etcd) systemd services.
* Installing [hyperkube](https://github.com/kubernetes/kubernetes/tree/master/cluster/images/hyperkube) that is used to set up Kubernetes cluster for experimantation purposes.
* Setting up SSH for root.
* Installing [Go](https://golang.org/).

### Provisioners

There are two provisioners defined in the [Vagrantfile](Vagrantfile): `aws` and `virtualbox`. Our CI infracture uses the first of them while the second should be used for development. If you try to use `aws` provider on your own it will fail as AMI is not publicly available.
