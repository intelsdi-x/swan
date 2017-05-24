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


SWAN_BIN=/opt/swan/bin
SWAN_VERSION="v0.13"

K8S_VERSION="v1.6.3"
SNAP_VERSION="1.2.0"
ETCD_VERSION="3.1.0"
DOCKER_VERSION="17.03.0.ce-1.el7.centos"
SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION=5
SNAP_PLUGIN_PROCESSOR_TAG_VERSION=3
SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION=6
SNAP_PLUGIN_PUBLISHER_FILE_VERSION=2
SNAP_PLUGIN_COLLECTOR_USE_VERSION=1

echo "------------------------ Install OS Packages (`date`)"
yum makecache fast -y -q
yum update -y -q
yum install -y -q epel-release  # Enables EPEL repo

yum install -y -q \
    wget \
    etcd-${ETCD_VERSION} \
    python-pip \
    python-devel \
    libcgroup-tools \
    boost \
    glog \
    protobuf \
    opencv \
    hdf5 \
    leveldb \
    lmdb \
    opencv \
    libgomp \
    libevent \
    git \
    zeromq \
    cppzmq-devel \
    gengetopt \
    libevent-devel \
    scons \
    gcc-c++

echo "------------------------ Prepare services (`date`)"
function daemonStatus() {
    echo "$1 service status: $(systemctl show -p SubState $1 | cut -d'=' -f2)"
}

echo "Reload services"
systemctl daemon-reload

echo "----------------------------- Install Docker (`date`)"
# https://docs.docker.com/engine/installation/linux/centos/#install-using-the-repository
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum makecache fast -y -q
yum install -y -q docker-ce-${DOCKER_VERSION}
echo "Restart docker"
systemctl enable docker
systemctl start docker
docker run hello-world
gpasswd -a $SWAN_USER docker
daemonStatus docker

echo "----------------------------- Create Swan Installation Directory (`date`)"
mkdir -p ${SWAN_BIN}


echo "----------------------------- Install Swan Release Package(`date`)"
wget --no-verbose https://github.com/intelsdi-x/swan/releases/download/${SWAN_VERSION}/swan.tar.gz -O /tmp/swan.tar.gz
tar -xzf /tmp/swan.tar.gz -C ${SWAN_BIN}


echo "----------------------------- Pulling docker image (`date`)"
docker pull intelsdi/swan


echo "----------------------------- Retrieve binares from Docker container (`date`)"
docker run --rm -v /opt:/output intelsdi/swan cp -R /opt/swan /output


echo "----------------------------- Install Kubernetes (`date`)"
wget --no-verbose https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/hyperkube -O ${SWAN_BIN}/hyperkube
chmod +x ${SWAN_BIN}/hyperkube


echo "----------------------------- Install etcd (`date`)"
systemctl enable etcd
systemctl restart etcd
daemonStatus etcd


echo "----------------------------- Install Cassandra (`date`)"
cp /vagrant/cassandra/cassandra.service /etc/systemd/system
mkdir -p /var/data/cassandra
chcon -Rt svirt_sandbox_file_t /var/data/cassandra # SELinux policy
systemctl enable cassandra
echo "Restart Cassandra"
systemctl restart cassandra
daemonStatus cassandra


echo "----------------------------- Install Snap telemetry (`date`)"
curl -s https://packagecloud.io/install/repositories/intelsdi-x/snap/script.rpm.sh | bash
yum list -q --show-duplicates snap-telemetry
yum install -y -q snap-telemetry-${SNAP_VERSION}
systemctl enable snap-telemetry
systemctl start snap-telemetry
daemonStatus snap-telemetry


echo "----------------------------- Install external Snap plugins (`date`)"
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-collector-docker/releases/download/${SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION}/snap-plugin-collector-docker_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-collector-docker
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-publisher-cassandra/releases/download/${SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION}/snap-plugin-publisher-cassandra_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-publisher-cassandra
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-processor-tag/releases/download/${SNAP_PLUGIN_PROCESSOR_TAG_VERSION}/snap-plugin-processor-tag_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-processor-tag
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/${SNAP_PLUGIN_PUBLISHER_FILE_VERSION}/snap-plugin-publisher-file_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-publisher-file
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-collector-use/releases/download/${SNAP_PLUGIN_COLLECTOR_USE_VERSION}/snap-plugin-collector-use_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-collector-use


echo "---------------------------- Post install (`date`)"
chmod +x -R /opt/swan/bin
chown -R $SWAN_USER:$SWAN_USER $HOME_DIR
chown -R $SWAN_USER:$SWAN_USER /opt/swan
chmod -R +x /opt/swan/bin/*
ln -svf ${SWAN_BIN}/* /bin/

echo "---------------------------- Provisioning experiment environment done (`date`)"
