#!/bin/bash

set -e

echo "Setting up environment..."
## Setting up envs
echo "export GOPATH=\"$HOME_DIR/go\"" >> $HOME_DIR/.bash_profile
echo "export CCACHE_CONFIGPATH=\"/etc/ccache.conf\"" >> $HOME_DIR/.bash_profile
echo 'export PATH="/usr/lib64/ccache/:$PATH:/usr/local/go/bin:$GOPATH/bin"' >> $HOME_DIR/.bash_profile
echo 'export ATHENA_DIR=$GOPATH/src/github.com/intelsdi-x/athena' >> $HOME_DIR/.bash_profile

## Create convenient symlinks in the home directory
ln -sf $HOME_DIR/go/src/github.com/intelsdi-x/swan $HOME_DIR
