#!/usr/bin/env bash
#
# Downloads and prepares CIFAR10 training workload.

cd ./caffe_src
if [ -e ./examples/cifar10/cifar10_train_lmdb/data.mdb ] && [ -e ./examples/cifar10/cifar10_test_lmdb/data.mdb ]
then
    echo "Skipping preparation of CIFAR10 dataset"
else
    ./data/cifar10/get_cifar10.sh
    ./examples/cifar10/create_cifar10.sh
fi
