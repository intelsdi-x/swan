#!/usr/bin/env bash

set -e -o pipefail

yum install -y epel-release
yum groupinstall -y 'Development Tools'

yum install -y protobuf protobuf-devel leveldb leveldb-devel snappy \
    snappy-devel opencv opencv-devel boost boost-devel hdf5 hdf5-devel \
    gflags glog lmdb gflags-devel glog-devel lmdb-devel

# OpenBLAS shoudl be used for local developement, when not using MKL
yum install -y openblas openblas-devel
