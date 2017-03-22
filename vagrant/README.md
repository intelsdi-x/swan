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
> make build_swan
```

## Setup
For details how to create a Linux virtual machine pre-configured for running the Swan experiment, please refer to [Installation guide for Swan](../../../../docs/install.md)

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
Vagrant will set 2 CPUs and 4096 MB RAM for VM by default. Developer can override these values with following environmental variables:
- `VBOX_CPUS` - ***Note: integration tests fail with less than 2***
- `VBOX_MEM` - ***Note: integration tests tend to crash with less (gcc)***

***Please be informed that every single glide operation inside VM might affect your host's ~/.glide.***
By default your local `~/.glide` cache will be used as glide cache inside VM.

## Troubleshooting
- The integration tests require cassandra to be running. In this
  environment, systemd is responsible for keeping it alive. You can see
  how it's doing by running `systemctl status cassandra` and
  `journalctl -fu cassandra`
- To re-run the VM provisioning shell script manually, do:
  `vagrant destroy -f && vagrant up --no-provision && vagrant ssh`
  `sudo /vagrant/provision.sh`

