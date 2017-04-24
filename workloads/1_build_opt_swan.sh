#!/bin/bash

set -e

# rm -rf opt/swan

mkdir -p opt/swan/{bin,share}

cp memcached/output/memcached opt/swan/bin
cp intel-cmt-cat/output/* opt/swan/bin
cp stress-ng/output/* opt/swan/bin

# Caffe
cp caffe/output/bin/* opt/swan/bin/
cp -r caffe/output/share/caffe opt/swan/share/

