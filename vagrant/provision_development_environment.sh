#!/usr/bin/env bash
# Copyright (c) 2017 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -x -e -o pipefail

if [ "$USER" != "root" ]; then
    echo "This script needs to be run with root privileges"
    exit 1
fi 
if [ "$SWAN_USER" == "" ]; then
    echo "You need to set SWAN_USER environmental variable"
    exit 1
fi
if [ "$HOME_DIR" == "" ]; then
    echo "You need to set HOME_DIR environmental variable"
    exit 1
fi


echo "---------------------- Start provisioning (`date`)"

GO_VERSION="1.10"

echo "-------------------------- Setup up environment (`date`)"
function addEnv() {
    grep "$1" $HOME_DIR/.bash_profile || echo "$1" >> $HOME_DIR/.bash_profile
}
addEnv "export GOPATH=\"$HOME_DIR/go\""
# jupyter integration tests from notebooks
addEnv 'export PYTHONPATH=$GOPATH/src/github.com/intelsdi-x/swan'
addEnv 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin'


echo "------------------------ Install packages (`date`)"
echo "workload build dependencies"
yum install -y -q \
    cppzmq-devel \
    gengetopt \
    libevent-devel \
    scons \
    gcc-c++

echo "developer tools & provisioning dependencies"
yum install -y -q \
    gcc \
    curl \
    vim \
    tmux \
    htop \
    sudo \
    git \
    nmap-ncat

echo "----------------------------- Preparing SSH access for root (`date`)"
# root user
mkdir -p /root/.ssh
# known hosts
touch /root/.ssh/known_hosts
ssh-keyscan github.com >> /root/.ssh/known_hosts
ssh-keyscan localhost >> /root/.ssh/known_hosts
ssh-keyscan 127.0.0.1 >> /root/.ssh/known_hosts
# Generate ssh keys for root - needed to run an experiment with remote ssh executor.
rm -rf /root/.ssh/id_rsa
ssh-keygen -f /root/.ssh/id_rsa -t rsa -N ''
cat /root/.ssh/id_rsa.pub >> /root/.ssh/authorized_keys
chmod og-wx /root/.ssh/authorized_keys


echo "--------------------------- Go language (`date`)"
echo "Download go"
GOTGZ=/tmp/go${GO_VERSION}.linux-amd64.tar.gz
wget --no-verbose https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz -O $GOTGZ

echo "Install go"
mkdir -p /usr/local
tar -C /usr/local -xzf $GOTGZ 
/usr/local/go/bin/go version


echo "--------------------------- Post install (`date`)"
ln -sf $HOME_DIR/go/src/github.com/intelsdi-x/swan $HOME_DIR
chown -R $SWAN_USER:$SWAN_USER $HOME_DIR
chown -R $SWAN_USER:$SWAN_USER /opt/swan

echo "--------------------------- Provisioning development environment done (`date`)"
