#!/bin/bash

set -e

CAFFE_ROOT_DIRECTORY="$(pwd)/$(dirname ${BASH_SOURCE[0]})"
OPENBLAS_SRC_DIRECTORY="${CAFFE_ROOT_DIRECTORY}/openblas"
OPENBLAS_BIN_DIRECTORY="${CAFFE_ROOT_DIRECTORY}/build"
CAFFE_SRC_DIRECTORY="${CAFFE_ROOT_DIRECTORY}/caffe_src"

CPUS_NUMBER=$(grep -c ^processor /proc/cpuinfo)

pushd ${CAFFE_ROOT_DIRECTORY}
if [ "${BUILD_OPENBLAS}" == "true" ]; then
    pushd ${OPENBLAS_SRC_DIRECTORY}
    mkdir -p ${OPENBLAS_BIN_DIRECTORY}
    make -j USE_OPENMP=1 --quiet libs
    make -j USE_OPENMP=1 --quiet netlib
    make -j USE_OPENMP=1 --quiet shared
    make PREFIX=${OPENBLAS_BIN_DIRECTORY} --quiet install
    popd
    cp ${CAFFE_ROOT_DIRECTORY}/Makefile.config_openblas ${CAFFE_SRC_DIRECTORY}/Makefile.config
    export LD_LIBRARY_PATH=${OPENBLAS_BIN_DIRECTORY}/lib:$LD_LIBRARY_PATH
else
    echo "To build multithreaded caffe you need to set \"BUILD_OPENBLAS\" envs first."
    cp ${CAFFE_ROOT_DIRECTORY}/Makefile.config ${CAFFE_SRC_DIRECTORY}/Makefile.config
fi

cp ${CAFFE_ROOT_DIRECTORY}/caffe_cpu_solver.patch ${CAFFE_SRC_DIRECTORY}
cp ${CAFFE_ROOT_DIRECTORY}/vagrant_vboxsf_workaround.patch ${CAFFE_SRC_DIRECTORY}
cp ${CAFFE_ROOT_DIRECTORY}/get_cifar10.patch ${CAFFE_SRC_DIRECTORY}


pushd ${CAFFE_SRC_DIRECTORY}
patch -p1 --forward -s --merge < caffe_cpu_solver.patch
patch -p1 --forward -s --merge < vagrant_vboxsf_workaround.patch
patch -p1 --forward -s < get_cifar10.patch || true
export OMP_NUM_THREADS=${CPUS_NUMBER}
make --quiet all
popd
popd
