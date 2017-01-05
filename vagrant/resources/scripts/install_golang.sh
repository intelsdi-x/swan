#!/bin/bash

set -e

source $HOME_DIR/.bash_profile

GO_VERSION="1.7"
GLIDE_VERSION="v0.12.2"
GODEP_VERSION="v74"

echo "Installing Go..."
if [ ! -f /cache/go${GO_VERSION}.linux-amd64.tar.gz ]; then
    wget -q -P /cache https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz
    rm -fr /usr/local/go
    tar xf /cache/go${GO_VERSION}.linux-amd64.tar.gz -C /usr/local
fi

echo "Installing Glide..."
if [ ! -f /cache/glide-${GLIDE_VERSION}-linux-amd64.tar.gz ]; then
    wget -q -P /cache https://github.com/Masterminds/glide/releases/download/${GLIDE_VERSION}/glide-${GLIDE_VERSION}-linux-amd64.tar.gz
    mkdir -p $GOPATH/bin || true
    tar -xf /cache/glide-${GLIDE_VERSION}-linux-amd64.tar.gz linux-amd64/glide --strip-components=1 -C $GOPATH/bin/glide
fi

echo "Installing Godep..."
if [ ! -f /cache/.godep_${GODEP_VERSION} ]; then
    wget -q -P /cache https://github.com/tools/godep/releases/download/${GODEP_VERSION}/godep_linux_amd64
    mkdir -p $GOPATH/bin || true
    chmod +x /cache/godep_linux_amd64
    cp /cache/godep_linux_amd64 $GOPATH/bin/godep
    touch /cache/.godep_${GODEP_VERSION}
fi
