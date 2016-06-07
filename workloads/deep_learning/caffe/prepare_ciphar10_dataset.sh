#!/usr/bin/env bash
#
# Downloads and prepares CIFAR10 training workload.

cd caffe_src
./data/cifar10/get_cifar10.sh
./examples/cifar10/create_cifar10.sh
