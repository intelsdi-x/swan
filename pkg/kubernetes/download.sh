#!/usr/bin/env bash

# NOTE: This is a script for manual deployment of k8s binaries. It should be replaced by Ansible
# scripts.
set -x

# Remove version which could be already there.
rm -rf kubernetes.tar.gz

# Download 1.3.0 k8s.
wget https://github.com/kubernetes/kubernetes/releases/download/v1.3.0/kubernetes.tar.gz

# Untar k8s.
tar -zxvf kubernetes.tar.gz

pushd kubernetes/server
    # Untar k8s binaries for amd64 systems.
    tar -xzvf kubernetes-server-linux-amd64.tar.gz

    chmod +x `pwd`/kubernetes/server/bin/*
    # Symlink required binaries with force flag.

    ln -fs `pwd`/kubernetes/server/bin/kubectl /usr/bin/
    ln -fs `pwd`/kubernetes/server/bin/kube-apiserver /usr/bin/
    ln -fs `pwd`/kubernetes/server/bin/kube-controller-manager /usr/bin/
    ln -fs `pwd`/kubernetes/server/bin/kube-scheduler /usr/bin/
    ln -fs `pwd`/kubernetes/server/bin/kube-proxy /usr/bin/
    ln -fs `pwd`/kubernetes/server/bin/kubelet /usr/bin/
popd