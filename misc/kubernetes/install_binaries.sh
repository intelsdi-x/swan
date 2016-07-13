#!/usr/bin/env bash

# NOTE: This is a script for manual deployment of k8s binaries. It should be replaced by Ansible
# scripts.
set -x

pushd `dirname $0`
    SWAN_ROOT=`pwd`/../../
    SWAN_BIN=${SWAN_ROOT}misc/bin/
    SWAN_LIB=${SWAN_ROOT}misc/lib/

    mkdir -p ${SWAN_BIN}
    mkdir -p ${SWAN_LIB}

    OPT=$1
    if [ "${OPT}" = "--force" ] || [ ! -d  ${SWAN_LIB}kubernetes ] ; then
        pushd ${SWAN_LIB}
            # Remove version which could be already there.
            rm -rf kubernetes.tar.gz

            # Download 1.3.0 k8s.
            wget https://github.com/kubernetes/kubernetes/releases/download/v1.3.0/kubernetes.tar.gz

            # Untar k8s.
            tar -zxvf kubernetes.tar.gz
        popd
    fi

    pushd ${SWAN_LIB}kubernetes/server
        # Untar k8s binaries for amd64 systems.
        tar -xzvf kubernetes-server-linux-amd64.tar.gz

        chmod +x `pwd`/kubernetes/server/bin/*

        # Symlink required binaries with force flag.
        ln -fs `pwd`/kubernetes/server/bin/kubectl ${SWAN_BIN}
        ln -fs `pwd`/kubernetes/server/bin/kube-apiserver ${SWAN_BIN}
        ln -fs `pwd`/kubernetes/server/bin/kube-controller-manager ${SWAN_BIN}
        ln -fs `pwd`/kubernetes/server/bin/kube-scheduler ${SWAN_BIN}
        ln -fs `pwd`/kubernetes/server/bin/kube-proxy ${SWAN_BIN}
        ln -fs `pwd`/kubernetes/server/bin/kubelet ${SWAN_BIN}
    popd
popd