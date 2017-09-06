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

# Ansible swan deployment playbook

## Introduction
To deploy your own swan experiments to remote nodes we provide a environment preparation scripts both for your host and remote machines. Note that currently host preparation script uses `apt` package manager so `Debian` or `Ubuntu` would be a great choice. Remote installation scripts(ansible) on the other hand, are written for `CentOS7`. 

## Usage
First, to prepare your own machine for development and build Swan locally run following:
```bash
git clone https://github.com/intelsdi-x/swan.git
cd swan/ansible
sudo ./prepare_host_env.sh
source ~/.bashrc # Make $GOPATH and $GOROOT variables active
```
After the script has successfully installed all dependencies and built Swan binaries add all addresses of machines on which you wish to run Swan experiments to `inventory/cluster` file. Then run:
```bash
ansible-playbook -i inventory/cluster centos_deploy_playbook.yml
```
If all of the tasks ended up with `SUCCESS` or `CHANGED` status, then nodes are ready to start Swan experiments.

To learn more about Swan experiments read following [README.md](../experiments/memcached-sensitivity-profile/README.md)