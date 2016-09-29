#!/usr/bin/env bash

# NOTE: This is a script for manual deployment of k8s binaries. It should be replaced by Ansible
# scripts.
set -x -e -o pipefail

K8S_VERSION="v1.4.0-alpha.2-serenity"

CACHE_DIRECTORY=/cache

if [ ! -d /cache ]; then
    mkdir ~/.cache || true
    CACHE_DIRECTORY=~/.cache
fi

function downloadK8s() {
    if [ ! -f ${CACHE_DIRECTORY}/.kube-services-${K8S_VERSION} ] || [ ! -f ${CACHE_DIRECTORY}/$1 ] ; then
        wget -q https://s3-us-west-2.amazonaws.com/intel-sdi.eo.swan.kubernetes/${1}.${K8S_VERSION} -O ${CACHE_DIRECTORY}/$1
    fi 
    cp ${CACHE_DIRECTORY}/$1 ${ATHENA_BIN}
}

pushd `dirname $0`
    ATHENA_ROOT=`pwd`/../../
    ATHENA_BIN=${ATHENA_ROOT}misc/bin

    mkdir -p ${ATHENA_BIN}

    OPT=$1
    if [ "${OPT}" = "--force" ] || [ ! -f  ${ATHENA_BIN}/.kube-services-${K8S_VERSION} ] ; then
        pushd ${ATHENA_BIN}
            downloadK8s kubectl
            downloadK8s kube-apiserver
            downloadK8s kube-controller-manager
            downloadK8s kube-proxy
            downloadK8s kube-scheduler
            downloadK8s kubelet
            chmod +x ./*
            touch .kube-services-${K8S_VERSION}
            touch ${CACHE_DIRECTORY}/.kube-services-${K8S_VERSION}
        popd
    fi
popd
