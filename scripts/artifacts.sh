#!/bin/bash

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
    install -D -m755 ./build/experiments/memcached/memcached-sensitivity-profile ${ARTIFACTS_PATH}/bin
    install -D -m755 ./build/experiments/specjbb/specjbb-sensitivity-profile ${ARTIFACTS_PATH}/bin

    # snap & plugins
    cp ${GOPATH}/bin/{snaptel,snapteld,snap-plugin-collector-caffe-inference,snap-plugin-collector-docker,snap-plugin-collector-mutilate,snap-plugin-collector-specjbb,snap-plugin-processor-tag,snap-plugin-publisher-cassandra,snap-plugin-publisher-file,snap-plugin-publisher-session-test} ${ARTIFACTS_PATH}/bin

    # kubernetes
    cp --no-dereference misc/bin/{apiserver,controller-manager,federation-apiserver,federation-controller-manager,hyperkube,kubectl,kubelet,proxy,scheduler} ${ARTIFACTS_PATH}/bin

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

    # install openblas
    if [ -d "./workloads/deep_learning/caffe/openblas/build" ]; then
        install -D -m644 ./workloads/deep_learning/caffe/openblas/build/lib/* "${ARTIFACTS_PATH}/lib"
    else
        echo "Openblas won't be included - caffe won't be multithreaded"
    fi

    # install caffe
    install -d ${ARTIFACTS_PATH}/share/caffe

    install -D -m755 ./workloads/deep_learning/caffe/caffe_wrapper.sh "${ARTIFACTS_PATH}/bin/caffe_wrapper.sh"
    install -D -m644 ./workloads/deep_learning/caffe/caffe_src/build/lib/* "${ARTIFACTS_PATH}/lib"
    install -D -m755 ./workloads/deep_learning/caffe/caffe_src/build/tools/caffe.bin "${ARTIFACTS_PATH}/share/caffe/bin/caffe"
    install -D -m755 ./workloads/deep_learning/caffe/caffe_src/data/cifar10/get_cifar10.sh "${ARTIFACTS_PATH}/share/caffe/data/cifar10/get_cifar10.sh"	
    install -D -m755 ./workloads/deep_learning/caffe/caffe_src/build/tools/compute_image_mean "${ARTIFACTS_PATH}/share/caffe/build/tools/compute_image_mean"
    install -D -m755 ./workloads/deep_learning/caffe/caffe_src/build/examples/cifar10/convert_cifar_data.bin "${ARTIFACTS_PATH}/share/caffe/build/examples/cifar10/convert_cifar_data.bin"
    install -D -m644 ./workloads/deep_learning/caffe/cifar10_quick_iter_5000.caffemodel.h5 "${ARTIFACTS_PATH}/share/caffe/cifar10_quick_iter_5000.caffemodel.h5"

    # 
    install -d ${ARTIFACTS_PATH}/share/caffe/examples/cifar10/
    install -D -m644 ./workloads/deep_learning/caffe/caffe_src/examples/cifar10/* "${ARTIFACTS_PATH}/share/caffe/examples/cifar10/" 

    tar -czf swan.tgz -C ${ARTIFACTS_PATH} .
}

function install_swan {
    PREFIX="/opt/swan/"
    mkdir -p $PREFIX
    tar xf $(cat ./latest_build) -C ${PREFIX}
    export LD_LIBRARY_PATH="${PREFIX}/lib":$LD_LIBRARY_PATH
    export PATH="${PREFIX}/bin":$PATH
    caffe_wrapper.sh init
}


#
# function download {
#     s3cmd sync -c ${S3_CREDS_LOCATION} --no-preserve s3://${BUCKET_NAME}/build_artifacts/latest_build ./latest_build
#     FNAME=$(cat ./latest_build)
#     s3cmd sync -c ${S3_CREDS_LOCATION} --no-preserve s3://${BUCKET_NAME}/build_artifacts/${FNAME} ./${FNAME}
# }
#
# function upload {
#     _check_s3_params
#
#     FNAME=$(cat ./latest_build)
#     s3cmd put -c ${S3_CREDS_LOCATION} --no-preserve ./${FNAME} s3://${BUCKET_NAME}/build_artifacts/
#     s3cmd put -c ${S3_CREDS_LOCATION} --no-preserve ./latest_build s3://${BUCKET_NAME}/build_artifacts/
# }

case ${1} in
    "dist" )
        dist ;;
    "install" )
        install_swan ;;
    "uninstall" )
        uninstall_swan ;;
    "download" )
        download ;;
    "upload" )
        upload ;;
esac
