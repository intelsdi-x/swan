# Local single-node development using Vagrant and Virtualbox

## Quick start

```sh
$ git clone git@github.com:intelsdi-x/swan.git
$ cd swan/misc/dev/vagrant/singlenode
$ ssh-add ~/.ssh/id_rsa  # if not already added to host ssh agent
$ vagrant plugin install vagrant-vbguest  # automatic guest additions
$ vagrant box update
$ vagrant up  # takes a few minutes
$ vagrant ssh
> cd swan
> make deps
> make
```

## Prerequisites

- [Vagrant](https://vagrantup.com)
- [Virtualbox](https://www.virtualbox.org/wiki/Downloads)
- Read access to the
  [snap-plugin-publisher-cassandra](https://github.com/intelsdi-x/snap-plugin-publisher-cassandra)
  repository.

## What's provided "out of the box"

- CentOS 7 virtual machine with the following additional software packages:
  - docker
  - gengetopt
  - git
  - golang 1.6
  - libcgroup-tools
  - libevent-devel
  - nmap-ncat
  - perf
  - scons
  - tree
  - vim
  - wget

## Notes

- The project directory is mounted in the guest file system: edit with your
  preferred tools in the host OS!

## Setup
1. Make sure you have the newest [vagrant](https://www.vagrantup.com/downloads.html) version with VirtualBox as provider
1. `vagrant plugin install vagrant-vbguest`
1. `vagrant reload`
1. `cd $GOPATH/github.com/intelsdi-x/swan/misc/dev/vagrant/singlenode`
1. `vagrant up`

## Running the integration tests

1. SSH into the VM: `vagrant ssh`
1. Change to the swan directory: `cd ~/swan`
1. Fetch swan dependencies: `make deps`
1. Change to the snap directory:
   `cd $GOPATH/src/github.com/intelsdi-x/snap`
1. Build snap: `make deps && make`
1. Change to the swan directory: `cd ~/swan`
1. Run the integration tests: `make integration_test`

## Troubleshooting
- Vagrant 1.8.4 and Virtualbox 5.1.X aren't compatible, Virtualbox 5.0.10
  works fine with this Vagrant version
- If you can't run `make deps` because of unauthorized error, make sure you don't
  have in gitconfig:
  `[url "https://"]
           insteadOf = git://`
  Warning: removing this will disable ssh-agent authorization and in effect private repositories like cassandra plugin will become inaccessible.
  Note: (if you using proxying) make sure that your proxy can handle ssh connections.
- The integration tests require cassandra to be running. In this
  environment, systemd is responsible for keeping it alive. You can see
  how it's doing by running `systemctl status cassandra` and
  `journalctl -fu cassandra`
- If you get permission errors when trying to run the integration tests,
  you may need to remove build artifacts first by running `make clean`.
- If you get Caffe errors during building workloads saying
  'libcaffe.o cant not find "xxx"', go into caffe_src dir and run 'make clean'
- If you get a connection error when attempting to SSH into the guest
  VM and you're behind a proxy you may need to add an override rule to ignore
  SSH traffic to localhost.
- To re-run the VM provisioning shell scripts, do
  `vagrant up --provision`
