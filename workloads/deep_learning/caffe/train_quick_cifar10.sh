#!/usr/bin/env bash
#
# Runs Caffe workload with CIFAR10 example solver.
# You need to prepare the workload using `prepare_ciphar10_dataset.sh` first.

cd $(dirname ${BASH_SOURCE[0]})/../share/caffe
./bin/caffe train --solver=examples/cifar10/cifar10_quick_solver.prototxt
./bin/caffe train --solver=examples/cifar10/cifar10_quick_solver_lr1.prototxt --snapshot=examples/cifar10/cifar10_quick_iter_4000.solverstate.h5