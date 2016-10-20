#!/bin/bash

CAFFE_ROOT_DIRECTORY=$(dirname ${BASH_SOURCE[0]})
OPENBLAS_SRC_DIRECTORY="${CAFFE_ROOT_DIRECTORY}/openblas"
CAFFE_SRC_DIRECTORY="${CAFFE_ROOT_DIRECTORY}/caffe_src"

CPUS_NUMBER=$(grep -c ^processor /proc/cpuinfo)

pushd ${CAFFE_ROOT_DIRECTORY}
if [ "${OPENBLAS_PATH}" != "" ] && [ "${BUILD_OPENBLAS}" == "true" ]; then
    sudo mkdir -p ${OPENBLAS_PATH}
    pushd ${OPENBLAS_SRC_DIRECTORY}
    make -j USE_OPENMP=1 --quiet libs
    sudo make PREFIX=${OPENBLAS_PATH} --quiet install
    popd
    cp ${CAFFE_ROOT_DIRECTORY}/Makefile.config_openblas ${CAFFE_SRC_DIRECTORY}/Makefile.config
    export LD_LIBRARY_PATH=${OPENBLAS_PATH}/lib
else
    echo "To build multithreaded caffe you need to set \"OPENBLAS_PATH\" and \"BUILD_OPENBLAS\" envs first."
    cp ${CAFFE_ROOT_DIRECTORY}/Makefile.config ${CAFFE_SRC_DIRECTORY}/Makefile.config
fi

cp ${CAFFE_ROOT_DIRECTORY}/caffe_cpu_solver.patch ${CAFFE_SRC_DIRECTORY}
cp ${CAFFE_ROOT_DIRECTORY}/vagrant_vboxsf_workaround.patch ${CAFFE_SRC_DIRECTORY}
cp ${CAFFE_ROOT_DIRECTORY}/get_cifar10.patch ${CAFFE_SRC_DIRECTORY}


pushd ${CAFFE_SRC_DIRECTORY}
patch -p1 --forward -s --merge < caffe_cpu_solver.patch
patch -p1 --forward -s --merge < vagrant_vboxsf_workaround.patch
patch -p1 --forward -s --merge < get_cifar10.patch
export OMP_NUM_THREADS=${CPUS_NUMBER}
make --quiet all
popd
${CAFFE_ROOT_DIRECTORY}/prepare_cifar10_dataset.sh
popd
