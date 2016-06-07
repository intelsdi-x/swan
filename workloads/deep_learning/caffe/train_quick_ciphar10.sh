#!/usr/bin/env bash
#
# Runs Caffe workload with CIFAR10 example solver.
# You need to prepare the workload using `prepare_ciphar10_dataset.sh` first.

cd caffe_src
./examples/cifar10/train_quick.sh
