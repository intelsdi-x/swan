#!/usr/bin/env bash

# NOTE: This is a script for manual deployment of k8s binaries. It should be replaced by Ansible
# scripts.
set -x -e -o pipefail

K8S_VERSION="v1.4.0-beta.0"

pushd `dirname $0`
    SWAN_ROOT=`pwd`/../../
    SWAN_BIN=${SWAN_ROOT}misc/bin/

    mkdir -p ${SWAN_BIN}

    OPT=$1
    if [ "${OPT}" = "--force" ] || [ ! -f  ${SWAN_BIN}.kube-services-${K8S_VERSION} ] ; then
        pushd ${SWAN_BIN}
            wget https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubectl
            wget https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kube-apiserver
            wget https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kube-controller-manager
            wget https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kube-proxy
            wget https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kube-scheduler
            wget https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubelet
            chmod +x ./*
            touch .kube-services-${K8S_VERSION}
        popd
    fi
popd
