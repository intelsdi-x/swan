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

## Updating AMI image
1. Run [`swan-integration`](https://private.ci.snap-telemetry.io/job/swan-integration/build) job with parameters.
  - Example parameters:
    - `repo_organization: intelsdi-x`
    - `repo_branch: master`
    - `rebase_on_master: true`
    - `CLEANUP: true`
    - `BUILD_CACHED_IMAGE: true` ***(required)***
    - `SWAN_AMI: <empty>`
2. When job has been finished copy AMI ID (it looks like: `ami-xxxxxxxx`).
3. Paste AMI ID in `Vagrantfile` (`aws.ami` parameter).
4. Commit & Push your change.

## Manually running provision scripts
- Before running provision scripts, import your private ssh key into your GitHub account.
- All scripts are stored in `/vagrant/resources/scripts`.
- To manually run provision scripts run `./enter_developer_mode.sh <private key location>`
- Scripts order:
  1. `copy_configuration.sh`
  2. `install_packages.sh`
  3. `setup_env.sh`
  4. `setup_git.sh`
  5. `setup_services.sh`
  6. `install_golang.sh`
  7. `install_snap_athena.sh`
  8. `post_install.sh`
  9. `install_project_deps.sh`
  10. `checker.sh`

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
