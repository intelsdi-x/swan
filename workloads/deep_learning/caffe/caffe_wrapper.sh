#!/bin/bash

SWAN_BIN=$(dirname ${BASH_SOURCE[0]})
cd $SWAN_BIN/../share/caffe
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$SWAN_BIN/../lib
if [ "$1" == "init" ]; then
        echo "Creating dataset"
        mkdir -p /tmp/caffe/examples/cifar10/
        bash ./data/cifar10/get_cifar10.sh
        bash ./examples/cifar10/create_cifar10.sh
        cp ./cifar10_quick_iter_5000.caffemodel.h5 /tmp/caffe
        # make sure that /tmp/caffe folder is accessible for other users (capital X for search only on directories)
        # caffe requires both executable and write permissions
        chmod -R a+rwX /tmp/caffe
        exit 0
fi
./bin/caffe "$@"
