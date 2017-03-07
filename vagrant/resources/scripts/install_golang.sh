#!/bin/bash

set -e

source $HOME_DIR/.bash_profile

GO_VERSION="1.7"
GLIDE_VERSION="v0.12.3"
GODEP_VERSION="v74"

echo "Installing Go..."
if [ ! -f /cache/go${GO_VERSION}.linux-amd64.tar.gz ]; then
    wget -q -P /cache https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz
    rm -fr /usr/local/go
    tar xf /cache/go${GO_VERSION}.linux-amd64.tar.gz -C /usr/local
fi
