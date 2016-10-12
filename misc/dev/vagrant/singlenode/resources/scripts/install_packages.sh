#!/bin/bash

set -e

echo "Installing centos packages..."
echo Updating package lists
yum makecache fast -y -q
yum update -y -q
yum install -y -q epel-release  # Enables EPEL repo
echo Installing packages
yum install -y -q \
    ccache \
    deltarpm \
    docker-engine \
    etcd \
    gcc-g++ \
    gengetopt \
    git \
    htop \
    iptables \
    libcgroup-tools \
    libevent-devel \
    moreutils-parallel \
    nmap-ncat \
    numactl \
    perf \
    psmisc \
    pssh \
    python-pip \
    python-devel \
    scons \
    sudo \
    tree \
    vim \
    wget \
    zeromq-devel
. $HOME_DIR/go/src/github.com/intelsdi-x/swan/workloads/deep_learning/caffe/caffe_deps_centos.sh
yum clean all
