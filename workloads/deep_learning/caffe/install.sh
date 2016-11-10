#!/bin/bash

if [ "$1" == "" ]; then
    echo "Give me some value here"
    exit 1
fi

ARTIFACTS_PATH=$1
pushd $(dirname ${BASH_SOURCE[0]})
    # mkdir -p ${ARTIFACTS_PATH}/lib/python2.7/site-packages/caffe/
    # cp -r python/caffe/* ${ARTIFACTS_PATH}/lib/python2.7/site-packages/caffe/
    # TODO verify this
    # find ${ARTIFACTS_PATH}/lib/python2.7/site-packages/caffe/ -type d -exec touch '{}'/__init__.py \;

    install -D -m644 caffe_src/build/lib/* "${ARTIFACTS_PATH}/lib"
    install -D -m755 train_quick_cifar10.sh "${ARTIFACTS_PATH}/bin/"
    install -D -m755 caffe_src/build/tools/caffe.bin "${ARTIFACTS_PATH}/share/caffe/bin/caffe"
    install -D -m755 caffe_src/build/tools/compute_image_mean "${ARTIFACTS_PATH}/bin/compute_image_mean"
    install -D -m755 caffe_src/build/examples/cifar10/convert_cifar_data.bin "${ARTIFACTS_PATH}/bin/convert_cifar_data.bin"

    install -d ${ARTIFACTS_PATH}/share/caffe
    install -D -m644 /tmp/caffe/* ${ARTIFACTS_PATH}/share/caffe

    install -d ${ARTIFACTS_PATH}/share/caffe/data
    install -D -m644 caffe_src/data/cifar10 ${ARTIFACTS_PATH}/share/caffe/data

    # install -D -m755 build/examples/cifar10/convert_cifar_data.bin "${ARTIFACTS_PATH}/bin/convert_cifar_data"
    install -d ${ARTIFACTS_PATH}/share/caffe/examples/cifar10/
    install -D -m644 caffe_src/examples/cifar10/* "${ARTIFACTS_PATH}/share/caffe/examples/cifar10/" 
    # install -D -m755 build/examples/mnist/convert_mnist_data.bin "${ARTIFACTS_PATH}/bin/convert_mnist_data"
    # install -D -m755 build/examples/siamese/convert_mnist_siamese_data.bin "${ARTIFACTS_PATH}/bin/convert_mnist_siamese_data"
    # install -D -m755 build/tools/extract_features.bin "${ARTIFACTS_PATH}/bin/extract_features"
popd