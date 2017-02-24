#!/bin/bash

set -e

function addEnv() {
    grep "$1" $HOME_DIR/.bash_profile || echo "$1" >> $HOME_DIR/.bash_profile
}

function addGlobalEnv() {
    grep "$1" /etc/environment || echo "$1" >> /etc/environment
}

echo "Setting up environment..."
## Setting up envs
addEnv "export GOPATH=\"$HOME_DIR/go\""
addEnv 'export CCACHE_CONFIGPATH=/etc/ccache.conf'
addEnv 'export SWAN_DIR=$GOPATH/src/github.com/intelsdi-x/swan'
addEnv 'export PYTHONPATH=$PYTHONPATH:$SWAN_DIR'
addEnv 'export OPENBLAS_PATH=/opt/OpenBLAS'
addEnv 'export LD_LIBRARY_PATH=$OPENBLAS_PATH/lib'
addEnv 'export PATH=/usr/lib64/ccache/:$PATH:/opt/swan/bin:/usr/local/go/bin:$GOPATH/bin'

## Create convenient symlinks in the home directory
ln -sf $HOME_DIR/go/src/github.com/intelsdi-x/swan $HOME_DIR

## Make sure that all required packages are also available for remote access. 
addGlobalEnv  'PATH=/usr/lib64/ccache:/sbin:/bin:/usr/sbin:/usr/bin:/opt/swan/bin'
