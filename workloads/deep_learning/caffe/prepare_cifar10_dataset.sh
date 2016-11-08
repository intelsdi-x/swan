#!/usr/bin/env bash
#
# Downloads and prepares CIFAR10 training workload.
# LMDB files are put outside Swan directory (/tmp/caffe). Vagrant has problem to support LMDB files over vboxsf (virtualbox.org/ticket/819).

cd ./caffe_src

if [ -e ./data/cifar10/batches.meta.txt ]
then
    echo "Skipping downloading of CIFAR10 dataset"
else
    ./data/cifar10/get_cifar10.sh
fi

if [ -e /tmp/caffe/examples/cifar10/cifar10_train_lmdb/data.mdb ] && [ -e /tmp/caffe/examples/cifar10/cifar10_test_lmdb/data.mdb ]
then
    echo "Skipping preparation of CIFAR10 dataset"
else
    echo "Preparing CIFAR10 dataset in /tmp/caffe/examples/cifar10/ directory"
    mkdir -p /tmp/caffe/examples/cifar10/
    ./examples/cifar10/create_cifar10.sh
fi

# ./examples/cifar10/create_cifar10.sh converts data from ./data/cifar10/* into 
# /tmp/caffe/examples/cifar10_[train|test]_lmdb database so downloaded files are not
# needed anymore.
if [ -e /tmp/caffe/examples/cifar10/cifar10_train_lmdb/data.mdb ] && [ -e /tmp/caffe/examples/cifar10/cifar10_test_lmdb/data.mdb ]
then
    echo "Removing downloaded data"
    rm -f ./data/cifar10/batches.meta.txt ./data/cifar10/data_batch_?.bin ./data/cifar10/readme.html ./data/cifar10/test_batch.bin
fi 

# Copy saved weight so we don't need to perform trainning.
if [ -e ../cifar10_quick_iter_5000.caffemodel.h5 ]
then
    cp ../cifar10_quick_iter_5000.caffemodel.h5 /tmp/caffe/
fi
