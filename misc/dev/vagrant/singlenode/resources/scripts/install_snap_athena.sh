#!/bin/bash

set -e

echo "prepare root to download athena"
mkdir -p ~/.ssh/
ssh-keyscan github.com >> ~/.ssh/known_hosts
git config --global url."git@github.com:".insteadOf "https://github.com/"

. $HOME_DIR/.bash_profile
ATHENA_DIR=$GOPATH/src/github.com/intelsdi-x/athena

echo "Installing Snap..."
if [ ! -f /cache/snap-v0.14.0-beta-linux-amd64.tar.gz ]; then
    wget -q -P /cache https://github.com/intelsdi-x/snap/releases/download/v0.14.0-beta/snap-v0.14.0-beta-linux-amd64.tar.gz
    tar xf /cache/snap-v0.14.0-beta-linux-amd64.tar.gz -C /cache
    mv /cache/snap-v0.14.0-beta/bin/* $GOPATH/bin
fi

echo "Installing Athena & its K8s..."
echo "Fetching Athena sources"
[ -d $ATHENA_DIR ] || (mkdir -p $ATHENA_DIR && git clone git@github.com:intelsdi-x/athena $ATHENA_DIR)
echo "Fetching kubernetes binaries for Athena"
cd $ATHENA_DIR && ./misc/kubernetes/install_binaries.sh
