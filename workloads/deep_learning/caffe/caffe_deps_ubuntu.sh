#!/usr/bin/env bash
#
# Prepares operating system for building Caffe Workload.
# Read $SWAN_ROOT/workloads/deep_learning/caffe/README.md for details.
set -e -o pipefail

apt-get install -y libprotobuf-dev libleveldb-dev \
    libsnappy-dev libopencv-dev libhdf5-serial-dev protobuf-compiler \
    libboost-all-dev libgflags-dev libgoogle-glog-dev liblmdb-dev \
    libopenblas-base libopenblas-dev
