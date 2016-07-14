# Installation guide for Swan

## Install OS dependencies

Swan is built to be run on Linux and has been tested on Linux Centos 7.

```
$ sudo yum update -y
$ sudo yum install -y epel-release
$ sudo yum clean all
$ yum install -y docker-engine gcc-g++ gengetopt git libcgroup-tools libevent-devel \
  nmap-ncat perf scons tree vim wget psmisc pssh
```

## Golang

A recent (1.6.x+) version of [Golang](https://golang.org/) is recommended. See [here](https://golang.org/doc/install) for guidance on installation of Golang. After installing Golang, make sure `$GOPATH` is set to point to the root of the workspace where the [Swan sources](https://github.com/intelsdi-x/swan) are checked out.

## Downloading the Swan sources

```
$ mkdir -p $GOPATH/src/github.com/intelsdi-x/swan
$ git clone https://github.com/intelsdi-x/swan.git $GOPATH/src/github.com/intelsdi-x/swan
```

## Install dependencies for Swan

```
$ cd $GOPATH/src/github.com/intelsdi-x/swan
$ make deps
$ pushd ../swan
$ make
$ popd
```

## Build experiment binary

```
$ make build
```

This will build and install the binaries in the `build/` directory

To run tests, please refer to the [Swan development guide](development.md).
