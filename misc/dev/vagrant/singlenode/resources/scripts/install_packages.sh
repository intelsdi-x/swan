#!/bin/bash

set -e

echo "Installing centos packages..."
echo Updating package lists
yum makecache fast -y -q
yum update -y -q
yum install -y -q epel-release  # Enables EPEL repo
echo Installing packages
yum groupinstall -y -q "Development tools"
yum install -y -q \
    boost \
    boost-devel \
    ccache \
    cppzmq-devel \
    deltarpm \
    docker-engine \
    etcd \
    gcc-g++ \
    gengetopt \
    gflags \
    gflags-devel \
    git \
    glog \
    glog-devel \
    hdf5 \
    hdf5-devel \
    hg \
    htop \
    iptables \
    java-1.8.0-openjdk \
    leveldb \
    leveldb-devel \
    libcgroup-tools \
    libevent-devel \
    lmdb \
    lmdb-devel \
    moreutils-parallel \
    nmap-ncat \
    numactl \
    openblas \
    openblas-devel \
    opencv \
    opencv-devel \
    perf \
    protobuf \
    protobuf-devel \
    psmisc \
    pssh \
    python-pip \
    python-devel \
    snappy \
    snappy-devel \
    scons \
    sudo \
    tree \
    vim \
    wget \
    zeromq-devel
yum clean all

echo "Installing python packages"
pip install s3cmd
