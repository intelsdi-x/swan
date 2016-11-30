#!/bin/bash
cd $(dirname ${BASH_SOURCE[0]})/../

function dist {
    ARTIFACTS_PATH="$(pwd)/artifacts/"
    mkdir -p ${ARTIFACTS_PATH}/{bin,lib}

    # install memcached
    install -D -m755 ./workloads/data_caching/memcached/memcached-1.4.25/build/memcached ${ARTIFACTS_PATH}/bin
    # install mutilate
    install -D -m755 ./workloads/data_caching/memcached/mutilate/mutilate ${ARTIFACTS_PATH}/bin
    # install low level aggressors
    install -D -m755 ./workloads/low-level-aggressors/{l1d,l1i,l3,memBw,stream.100M} ${ARTIFACTS_PATH}/bin
    # copy go binaries
    cp ${GOPATH}/bin/* ${ARTIFACTS_PATH}/bin
    install -D -m755 ./build/experiments/memcached/memcached-sensitivity-profile ${ARTIFACTS_PATH}/bin
    install -D -m755 ./build/experiments/specjbb/specjbb-sensitivity-profile ${ARTIFACTS_PATH}/bin

    # install specjbb
    install -d ${ARTIFACTS_PATH}/share/specjbb
    install -d ${ARTIFACTS_PATH}/share/specjbb/config
    install -d ${ARTIFACTS_PATH}/share/specjbb/lib

    if [ -d "./workloads/web_serving/specjbb/" ]; then
        install -D -m755 ./workloads/web_serving/specjbb/specjbb2015.jar ${ARTIFACTS_PATH}/share/specjbb/specjbb2015.jar
        install -D -m644 ./workloads/web_serving/specjbb/config/* ${ARTIFACTS_PATH}/share/specjbb/config/
        install -D -m644 ./workloads/web_serving/specjbb/lib/*.jar ${ARTIFACTS_PATH}/share/specjbb/lib/
    else
        echo "Specjbb won't be included"
    fi

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
    FILENAME="swan-$(date +%m%d%y-%H%M).tar.gz"
    echo ${FILENAME} > ./latest_build

    tar -czf ${FILENAME} -C ${ARTIFACTS_PATH} .
}

function install_swan {
    if [ "${UID}" != 0 ]; then
        >&2 echo "Only root can perform this operation"
        exit 1
    fi

    if [ "${PREFIX}" == "" ]; then
        PREFIX="/usr/"
    fi

    tar xf $(cat ./latest_build) -C ${PREFIX}
    export LD_LIBRARY_PATH="${PREFIX}/lib":$LD_LIBRARY_PATH
    export PATH="${PREFIX}/bin":$PATH
    caffe init
}

function uninstall_swan {
    if [ "${UID}" != 0 ]; then
        >&2 echo "Only root can perform this operation"
        exit 1
    fi

    if [ "${PREFIX}" == "" ]; then
        PREFIX="/usr/"
    fi

    FILELIST=$(tar ztf $(cat ./latest_build))
    for listed_file in $FILELIST; do
        if [[ -f ${PREFIX}/${listed_file} ]]; then
            rm -v ${PREFIX}/${listed_file}
        fi
    done
}

case ${1} in
    "dist" )
        dist ;;
    "install" )
        install_swan ;;
    "uninstall" )
        uninstall_swan ;;
esac
