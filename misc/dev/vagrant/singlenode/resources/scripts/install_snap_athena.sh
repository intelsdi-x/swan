#!/bin/bash

set -e

SNAP_VERSION="v1.0.0"

. $HOME_DIR/.bash_profile
ATHENA_DIR=$GOPATH/src/github.com/intelsdi-x/athena

echo "Installing Snap..."
if [ ! -f /cache/snap-${SNAP_VERSION}-linux-amd64.tar.gz ]; then
    wget -q -P /cache https://github.com/intelsdi-x/snap/releases/download/${SNAP_VERSION}/snap-${SNAP_VERSION}-linux-amd64.tar.gz
    tar xf /cache/snap-${SNAP_VERSION}-linux-amd64.tar.gz -C /cache
    mv /cache/snap-${SNAP_VERSION}/bin/* $GOPATH/bin
fi

echo "Installing Athena & its K8s..."
if [ ! -d $ATHENA_DIR ]; then
    echo "Fetching Athena sources"
    mkdir -p $ATHENA_DIR 
    git clone git@github.com:intelsdi-x/athena $ATHENA_DIR
else
    echo "Updating Athena sources"
    pushd $ATHENA_DIR
    git pull
    popd
fi
echo "Fetching kubernetes binaries for Athena"
cd $ATHENA_DIR && ./misc/kubernetes/install_binaries.sh
