#!/bin/bash -x

CAFFE_ROOT_DIRECTORY=$(dirname ${BASH_SOURCE[0]})
OPENBLAS_SRC_DIRECTORY="${CURRENT_DIRECTORY}\openblas"
CAFFE_SRC_DIRECTORY="${CURRENT_DIRECTORY}\caffe_src"

CPUS_NUMBER=$(grep -c ^processor /proc/cpuinfo)


pushd ${CAFFE_ROOT_DIRECTORY}
if [ "${OPENBLAS_PATH}" == "true" ]; then
    pushd ${OPENBLAS_SRC_DIRECTORY}
    make -j USE_OPENMP=1
    make PREFIX=${OPENBLAS_PATH} install
    popd
    cp ${CAFFE_ROOT_DIRECTORY}/Makefile.config_openblas ${CAFFE_SRC_DIRECTORY}/Makefile.config
    export LD_LIBRARY_PATH=${OPENBLAS_PATH}/lib
else
    echo "To build multithreaded caffe you need to set \"OPENBLAS_PATH\" env first."
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
make all
popd
${CAFFE_ROOT_DIRECTORY}/prepare_cifar10_dataset.sh
popd