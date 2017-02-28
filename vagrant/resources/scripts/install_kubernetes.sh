#!/usr/bin/env bash

# NOTE: This is a script for manual deployment of k8s binaries. It should be replaced by Ansible
# scripts.
set -x -e -o pipefail

source $HOME_DIR/.bash_profile

K8S_VERSION="v1.5.1"

CACHE_DIRECTORY=/cache

if [ ! -d /cache ]; then
    mkdir -p ~/.cache
    CACHE_DIRECTORY=~/.cache
fi

pushd `dirname $0`
    SWAN_BIN=${SWAN_DIR}/misc/bin

    mkdir -p ${SWAN_BIN}

    if [ ! -f ${CACHE_DIRECTORY}/.kube-services-${K8S_VERSION} ]; then
        # instead of downloading multiple binaries only hyperkube is downloaded
        wget -q https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/hyperkube -O ${CACHE_DIRECTORY}/hyperkube-${K8S_VERSION}
        chmod +x ${CACHE_DIRECTORY}/hyperkube-${K8S_VERSION}

        touch ${CACHE_DIRECTORY}/.kube-services-${K8S_VERSION}
    fi
    # to make usage easier - symlinks are generated for hyperkube in PATH
    cp ${CACHE_DIRECTORY}/hyperkube-${K8S_VERSION} ${SWAN_BIN}/hyperkube
    pushd ${SWAN_BIN}
    ./hyperkube --make-symlinks || true # ignore any errors  (like existing symlinks)
    popd
popd
