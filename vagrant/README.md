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

First - install: [Vagrant](https://www.vagrantup.com/docs/installation/), [VirtualBox](https://www.virtualbox.org/wiki/Downloads) and [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git). Then you just need to execute following commands (they should work on any flavour of Linux):

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

Now you should be able to run experiment on automatically provisioned Kubernetes cluster. If you want to be able to run experiments without Kubernetes then you should proceed with further steps.

Some of the project dependencies are distributed as [Dcoker image](https://hub.docker.com/r/intelsdi/swan/). They need to be extracted from the container in order to be used for non-Kubernetes experiment. To do this you need to execute following commands:

```sh
make extract_binaries
sudo chown -R $USER:$USER opt 
sudo cp -fa opt/swan /opt
sudo ln -svf /opt/swan/bin/* /usr/bin/
```

## Troubleshooting

You may encounter follwing error:
```
Unable to export dependencies to vendor directory: remove /home/vagrant/swan/vendor/golang.org/x/sys/unix: directory not empty
```
You should remove `vendor` catalog completely in this case.

## Tuning VM parameters

Vagrant will allocate 2 CPUs and 4096 MB RAM for the VM by default. You can consider changing these values but:
* You need to provide at least 2 CPUs.
* You need to provide at least 4096 MB of RAM.

## Deeper dive

If you want to learn more about VM configuration and installed packages refer to [provisioning script](provision.sh).

`~/.glide` directory from your host will be mounted on the VM to speed up Go dependency management.

The script is responsible for:
* Installing all the necessary packages that are needed to build Swan, run experiments and analyse their results.
* Installing [Snap](http://snap-telemetry.io/) and its plugins that are responsible for gathering experiment results.
* Installing [Docker](https://www.docker.com/) that allows running experiment on [Kubernetes](https://kubernetes.io) cluster.
* Enabling [Cassandra](http://cassandra.apache.org/) and [Etcd](https://coreos.com/etcd) systemd services.
* Installing [hyperkube](https://github.com/kubernetes/kubernetes/tree/master/cluster/images/hyperkube) that is used to set up Kubernetes cluster for experimantation purposes.
* Setting up SSH for root.
* Installing [Go](https://golang.org/).

### Privisioners

There are two provisioners defined in [Vagrantfile](Vagrantfile): `aws` and `virtualbox`. CI infracture uses first of them while the second should be used for development. If you try to use `aws` provider on your own it will fail as AMI is not publicly available.
