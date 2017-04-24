#!/bin/bash
cd $(dirname ${BASH_SOURCE[0]})
./caffe.sh test -model ./examples/cifar10/cifar10_quick_train_test.prototxt -weights ./examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5 -iterations 100000

