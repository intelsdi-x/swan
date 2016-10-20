#!/bin/bash

set -e

function addEnv() {
    grep "$1" $HOME_DIR/.bash_profile || echo "$1" >> $HOME_DIR/.bash_profile
}

echo "Setting up environment..."
## Setting up envs
addEnv "export GOPATH=\"$HOME_DIR/go\""
addEnv 'export CCACHE_CONFIGPATH=/etc/ccache.conf'
addEnv 'export PATH=/usr/lib64/ccache/:$PATH:/usr/local/go/bin:$GOPATH/bin'
addEnv 'export ATHENA_DIR=$GOPATH/src/github.com/intelsdi-x/athena'
addEnv 'export OPENBLAS_PATH=/opt/OpenBLAS'
addEnv 'export LD_LIBRARY_PATH=/opt/OpenBLAS/lib'

## Create convenient symlinks in the home directory
ln -sf $HOME_DIR/go/src/github.com/intelsdi-x/swan $HOME_DIR
