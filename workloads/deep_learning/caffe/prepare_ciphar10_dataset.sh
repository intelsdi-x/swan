#!/usr/bin/env bash

cd caffe_src
./data/cifar10/get_cifar10.sh
./examples/cifar10/create_cifar10.sh
