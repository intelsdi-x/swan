#!/bin/sh

cd $(dirname ${BASH_SOURCE[0]})/../share/caffe
if [ "$1" == "init" ]; then
        echo "Creating dataset"
        mkdir -p /tmp/caffe/examples/cifar10/
        bash ./data/cifar10/get_cifar10.sh
        bash ./examples/cifar10/create_cifar10.sh
        cp ./cifar10_quick_iter_5000.caffemodel.h5 /tmp/caffe
        exit 0
fi
./bin/caffe "$@"
