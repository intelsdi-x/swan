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

K8S_VERSION="v1.5.6"
SNAP_VERSION="1.2.0"
ETCD_VERSION="3.1.0"
DOCKER_VERSION="17.03.0.ce-1.el7.centos"
SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION=5
SNAP_PLUGIN_PROCESSOR_TAG_VERSION=3
SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION=5
SNAP_PLUGIN_PUBLISHER_FILE_VERSION=2
SNAP_PLUGIN_COLLECTOR_USE_VERSION=1

echo "------------------------ Install packages (`date`)"
yum makecache fast -y -q
yum update -y -q
yum install -y -q epel-release  # Enables EPEL repo

echo "swan dependecies"
yum install -y -q \
    python-pip \
    python-devel \
    etcd-${ETCD_VERSION} \
    libcgroup-tools \
    numactl \
    moreutils-parallel \
    nmap-ncat \
    wget

echo "workload runtime depedencies"
yum install -y -q \
    glog \
    protobuf \
    boost \
    hdf5 \
    leveldb \
    lmdb \
    opencv \
    libgomp \
    numactl-libs \
    libevent \
    zeromq \
    java-1.8.0-openjdk

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
# Install into /opt/swan/bin.
mkdir -p ${SWAN_BIN}
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-collector-docker/releases/download/${SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION}/snap-plugin-collector-docker_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-collector-docker
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-publisher-cassandra/releases/download/${SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION}/snap-plugin-publisher-cassandra_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-publisher-cassandra
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-processor-tag/releases/download/${SNAP_PLUGIN_PROCESSOR_TAG_VERSION}/snap-plugin-processor-tag_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-processor-tag
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/${SNAP_PLUGIN_PUBLISHER_FILE_VERSION}/snap-plugin-publisher-file_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-publisher-file
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-collector-use/releases/download/${SNAP_PLUGIN_COLLECTOR_USE_VERSION}/snap-plugin-collector-use_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-collector-use

chmod +x ${SWAN_BIN}/snap-plugin-collector-docker
chmod +x ${SWAN_BIN}/snap-plugin-publisher-cassandra
chmod +x ${SWAN_BIN}/snap-plugin-processor-tag
chmod +x ${SWAN_BIN}/snap-plugin-publisher-file
chmod +x ${SWAN_BIN}/snap-plugin-collector-use


echo "----------------------------- Install Kubernetes (`date`)"
wget --no-verbose https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/hyperkube -O ${SWAN_BIN}/hyperkube
chmod +x ${SWAN_BIN}/hyperkube
pushd ${SWAN_BIN}
    ./hyperkube --make-symlinks 
popd

echo "---------------------------- Post install (`date`)"
chown -R $SWAN_USER:$SWAN_USER $HOME_DIR
chown -R $SWAN_USER:$SWAN_USER /opt/swan

echo "---------------------------- Provisioning experiment environment done (`date`)"
