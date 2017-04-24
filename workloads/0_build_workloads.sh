#!/bin/bash

for workload in caffe memcached intel-cmt-cat stress-ng
do
    echo "Building $workload..."
    pushd $workload
    ./build.sh
    echo "$workload done"
    popd
done
