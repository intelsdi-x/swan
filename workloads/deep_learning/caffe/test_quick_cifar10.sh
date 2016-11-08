#!/bin/bash

# TBD: Missing license

CAFFETOOL=./build/tools/caffe
MODEL=./examples/cifar10/cifar10_quick_train_test.prototxt
WEIGHTS=/tmp/caffe/cifar10_quick_iter_5000.caffemodel.h5
ITERATIONS=1000000000
SIGINT=stop

if [ ! -e ${WEIGHTS} ]
then
    echo "Missing trained caffe model! Expected: ${WEIGHTS}"
    exit
fi

cd $(dirname ${BASH_SOURCE[0]})/caffe_src
exec ${CAFFETOOL} test -model ${MODEL} -weights ${WEIGHTS} -iterations ${ITERATIONS} -sigint_effect ${SIGINT}
