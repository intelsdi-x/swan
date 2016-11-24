#!/bin/bash

ARTIFACTS_PATH="$(pwd)/artifacts/"
mkdir -p $(ARTIFACTS_PATH)/{bin,lib}

# install memcached
install -D -m755 ./workloads/data_caching/memcached/memcached-1.4.25/build/memcached ${ARTIFACTS_PATH}/bin
# install mutilate
install -D -m755 ./workloads/data_caching/memcached/mutilate/mutilate ${ARTIFACTS_PATH}/bin
# install low level aggressors
install -D -m755 ./workloads/low-level-aggressors/{l1d,l1i,l3,memBw,stream.100M} ${ARTIFACTS_PATH}/bin
# copy go binaries
cp ${GOPATH}/bin/* ${ARTIFACTS_PATH}/bin

# install specjbb
install -d ${ARTIFACTS_PATH}/share/specjbb
install -d ${ARTIFACTS_PATH}/share/specjbb/config
install -d ${ARTIFACTS_PATH}/share/specjbb/lib
install -D -m755 ./workloads/web_serving/specjbb/specjbb2015.jar ${ARTIFACTS_PATH}/share/specjbb/specjbb2015.jar
install -D -m644 ./workloads/web_serving/specjbb/config/* ${ARTIFACTS_PATH}/share/specjbb/config/
install -D -m644 ./workloads/web_serving/specjbb/lib/*.jar ${ARTIFACTS_PATH}/share/specjbb/lib/

# install caffe
install -d ${ARTIFACTS_PATH}/share/caffe
install -D -m755 ./workloads/deep_learning/caffe/caffe_wrapper.sh "${ARTIFACTS_PATH}/bin/caffe"

install -D -m644 ./workloads/deep_learning/caffe/caffe_src/build/lib/* "${ARTIFACTS_PATH}/lib"

install -D -m755 ./workloads/deep_learning/caffe/caffe_src/build/tools/caffe.bin "${ARTIFACTS_PATH}/share/caffe/bin/caffe"
install -D -m755 ./workloads/deep_learning/caffe/caffe_src/data/cifar10/get_cifar10.sh "${ARTIFACTS_PATH}/share/caffe/data/cifar10/get_cifar10.sh"	
install -D -m755 ./workloads/deep_learning/caffe/caffe_src/build/tools/compute_image_mean "${ARTIFACTS_PATH}/share/caffe/build/tools/compute_image_mean"
install -D -m755 ./workloads/deep_learning/caffe/caffe_src/build/examples/cifar10/convert_cifar_data.bin "${ARTIFACTS_PATH}/share/caffe/build/examples/cifar10/convert_cifar_data.bin"
install -D -m644 ./workloads/deep_learning/caffe/cifar10_quick_iter_5000.caffemodel.h5 "${ARTIFACTS_PATH}/share/caffe/cifar10_quick_iter_5000.caffemodel.h5"

install -d ${ARTIFACTS_PATH}/share/caffe/examples/cifar10/
install -D -m644 ./workloads/deep_learning/caffe/caffe_src/examples/cifar10/* "${ARTIFACTS_PATH}/share/caffe/examples/cifar10/" 

# pack
tar -czf swan-$(date +%m%d%y-%H%M).tar.gz -C ${ARTIFACTS_PATH} .
