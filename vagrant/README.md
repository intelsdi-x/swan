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

```sh
$ git clone git@github.com:intelsdi-x/swan.git
$ cd swan/vagrant
$ vagrant plugin install vagrant-vbguest  # automatic guest additions
$ vagrant box update
$ vagrant up
$ vagrant ssh
> cd swan
> make deps build dist install
```

## Setup
For details on how to create a Linux virtual machine pre-configured for running the Swan experiment, please refer to [Installation guide for Swan](docs/install.md)

## Updating AMI image
1. Run [`swan-integration`](https://private.ci.snap-telemetry.io/job/swan-integration/build) job with parameters.
  - Example parameters:
    - `repo_organization: intelsdi-x`
    - `repo_branch: master`
    - `rebase_on_master: true`
    - `CLEANUP: true`
    - `BUILD_CACHED_IMAGE: true` ***(required)***
    - `SWAN_AMI: ami-6d1c2007`
2. When job has been finished copy AMI ID (it looks like: `ami-xxxxxxxx`).
3. Paste AMI ID in `Vagrantfile` (`aws.ami` parameter).
4. Commit & Push your change.

## Changing VM parameters
### Building additional artifacts
Depending on provider Vagrant may build a docker image and multithreaded caffe:
- For `aws` provider, vagrant won't build them by default - it's not necessary due to AMI caching
- For `virtualbox` provider, vagrant will build all of artifacts by default .

### VirtualBox CPUs and RAM values
Vagrant will set 2 CPUs and 4096 MB RAM for VM by default. Developer can override these values with the following environmental variables:
- `VBOX_CPUS` - ***Note: integration tests fail with less than 2***
- `VBOX_MEM` - ***Note: integration tests tend to crash with less (gcc)***

***WARNING: Please be informed that every single glide operation inside the VM might affect your host's ~/.glide.***
By default your local `~/.glide` cache will be used as your glide cache inside VM.

## Troubleshooting
- The integration tests require Cassandra to be running. In this
  environment, systemd is responsible for keeping it alive. You can see
  how it's doing by running `systemctl status cassandra` and
  `journalctl -fu cassandra`
- To re-run the VM provisioning shell script manually, do:
  `vagrant destroy -f && vagrant up --no-provision && vagrant ssh`
  `sudo /vagrant/provision.sh`
