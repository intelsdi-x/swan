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

## What's provided "out of the box"

- CentOS 7 virtual machine with the following additional software packages:
  - docker
  - gengetopt
  - git
  - golang 1.6
  - libcgroup-tools
  - libevent-devel
  - perf
  - scons
  - tree
  - vim
  - wget

## Notes

- The project directory is mounted in the guest file system: edit with your
  preferred tools in the host OS!

## Running the integration tests

1. SSH into the VM: `vagrant ssh`
1. Change to the swan directory: `cd ~/swan`
1. Fetch swan dependencies: `make deps`
1. Build snap:
  - `cd $GOPATH/src/github.com/intelsdi-x/snap`
  - `make`
1. Run the integration tests: `~/scripts/run-integration-tests.sh`

## Troubleshooting

- If you get a connection error when attempting to SSH into the guest
  VM and you're behind a proxy you may need to add an override rule to ignore
  SSH traffic to localhost.
- To re-run the VM provisioning shell scripts, do
  `vagrant up --provision`
