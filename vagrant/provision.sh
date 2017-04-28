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
echo "---------------------- Start provisioning (`date`)"

GO_VERSION="1.7.5"
K8S_VERSION="v1.5.6"
SNAP_VERSION="1.2.0"
ETCD_VERSION="3.1.0"
DOCKER_VERSION="17.03.0.ce-1.el7.centos"
SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION=5
SNAP_PLUGIN_COLLECTOR_USE_VERSION=1
SNAP_PLUGIN_PROCESSOR_TAG_VERSION=3
SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION=5
SNAP_PLUGIN_PUBLISHER_FILE_VERSION=2

# use some sane defaults if it run manually (please run as root)
HOME_DIR="${HOME_DIR:-/home/vagrant}"
VAGRANT_USER="${VAGRANT_USER:-vagrant}"
SWAN_BIN=/opt/swan/bin


echo "-------------------------- Setup up environment (`date`)"
function addEnv() {
    grep "$1" $HOME_DIR/.bash_profile || echo "$1" >> $HOME_DIR/.bash_profile
}
addEnv "export GOPATH=\"$HOME_DIR/go\""
# jupyter intergration tests from notebooks
addEnv 'export PYTHONPATH=$GOPATH/src/github.com/intelsdi-x/swan'
addEnv 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin'


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
    nmap-ncat

echo "workload depedencies"
yum install -y -q \
    glog protobuf boost hdf5 leveldb lmdb opencv libgomp numactl-libs \
    libevent zeromq java-1.8.0-openjdk-devel \
    java-1.8.0-openjdk-devel

echo "developer tools & provisioning depedencies"
yum install -y -q \
    gcc \
    curl \
    wget \
    vim \
    tmux \
    htop \
    sudo \
    git


echo "------------------------ Prepare services (`date`)"
function daemonStatus() {
    echo "$1 service status: $(systemctl show -p SubState $1 | cut -d'=' -f2)"
}

echo "Reload services"
systemctl daemon-reload


echo "Install docker"
# https://docs.docker.com/engine/installation/linux/centos/#install-using-the-repository
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum makecache fast -y -q
yum install -y -q docker-ce-${DOCKER_VERSION}
echo "Restart docker"
systemctl enable docker
systemctl start docker
docker run hello-world
gpasswd -a $VAGRANT_USER docker
daemonStatus docker

echo "Restart etcd" 
systemctl enable etcd
systemctl restart etcd
daemonStatus etcd

echo "Install Cassandra"
cp /vagrant/cassandra/cassandra.service /etc/systemd/system
mkdir -p /var/data/cassandra
chcon -Rt svirt_sandbox_file_t /var/data/cassandra # SELinux policy
systemctl enable cassandra
echo "Restart Cassandra"
systemctl restart cassandra
daemonStatus cassandra


echo "----------------------------- Install snap telemetry (`date`)"
curl -s https://packagecloud.io/install/repositories/intelsdi-x/snap/script.rpm.sh | bash
yum list -q --show-duplicates snap-telemetry
yum install -y -q snap-telemetry-${SNAP_VERSION}
systemctl enable snap-telemetry
systemctl start snap-telemetry
daemonStatus snap-telemetry


echo "----------------------------- Install external snap plugins (`date`)"
# Install into /opt/swan/bin.
mkdir -p ${SWAN_BIN}
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-collector-docker/releases/download/${SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION}/snap-plugin-collector-docker_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-collector-docker
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-collector-use/releases/download/${SNAP_PLUGIN_COLLECTOR_USE_VERSION}/snap-plugin-collector-use_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-collector-use
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-publisher-cassandra/releases/download/${SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION}/snap-plugin-publisher-cassandra_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-publisher-cassandra
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-processor-tag/releases/download/${SNAP_PLUGIN_PROCESSOR_TAG_VERSION}/snap-plugin-processor-tag_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-processor-tag
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/${SNAP_PLUGIN_PUBLISHER_FILE_VERSION}/snap-plugin-publisher-file_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-publisher-file

chmod +x ${SWAN_BIN}/snap-plugin-collector-docker
chmod +x ${SWAN_BIN}/snap-plugin-collector-use
chmod +x ${SWAN_BIN}/snap-plugin-publisher-cassandra
chmod +x ${SWAN_BIN}/snap-plugin-processor-tag
chmod +x ${SWAN_BIN}/snap-plugin-publisher-file


echo "----------------------------- Kubernetes (`date`)"
wget --no-verbose https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/hyperkube -O ${SWAN_BIN}/hyperkube
chmod +x ${SWAN_BIN}/hyperkube
pushd ${SWAN_BIN}
    ./hyperkube --make-symlinks 
popd


echo "----------------------------- Preparing SSH access for root (`date`)"
# root user
mkdir -p /root/.ssh
# known hosts
touch /root/.ssh/known_hosts
ssh-keyscan github.com >> /root/.ssh/known_hosts
ssh-keyscan localhost >> /root/.ssh/known_hosts
ssh-keyscan 127.0.0.1 >> /root/.ssh/known_hosts
# Generte ssh keys for root - needed to run an experiment with remote ssh executor.
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

##################################
# PRIVATE PART
##################################
# Enable this to have access to private repos.
#git config --global url."git@github.com:".insteadOf "https://github.com/"
echo "-------------------------- Private components (`date`)"
if [ -e "$HOME_DIR/swan_s3_creds/.s3cfg" ]; then
    cp $HOME_DIR/swan_s3_creds/.s3cfg ~/.s3cfg

    echo "Install s3cmd"
    pip install s3cmd

    echo "Install public keys"
    s3cmd get s3://swan-artifacts/public_keys authorized_keys
    cat authorized_keys >> ${HOME_DIR}/.ssh/authorized_keys
    
    echo "Install glide cache"
    if [ ! -d ${HOME_DIR}/.glide ]; then
        s3cmd get s3://swan-artifacts/glide-cache-2017-03-10.tgz /tmp/glide-cache.tgz
        tar --strip-components 2 -C ${HOME_DIR} -xzf /tmp/glide-cache.tgz
        chown -R ${VAGRANT_USER}:${VAGRANT_USER} ${HOME_DIR}/.glide
    fi

    echo "Synchronize /opt/swan"
    # For manuall installtion
    # sudo sh -c "mkdir -p /opt/swan && chown -R $USER:$USER /opt/swan"
    s3cmd sync s3://swan-artifacts/workloads/swan/ /opt/swan/

    echo "Install binaries from /opt/swan"
    ln -sv ${SWAN_BIN}/* /bin/

    echo "Download centos_swan_image docker image"
    s3cmd sync s3://swan-artifacts/workloads/centos_swan_image.tgz /tmp/centos_swan_image.tgz

    
    echo "Import centos_swan_image to docker"
    gunzip /tmp/centos_swan_image.tgz 
    docker image load -i /tmp/centos_swan_image.tar 
    docker images centos_swan_image

    echo "Check centos_swan_image docker"
    docker run --rm centos_swan_image memcached -V
    docker run --rm centos_swan_image mutilate --version
    docker run --rm centos_swan_image caffe.sh --version
fi

echo "--------------------------- Post install (`date`)"
ln -sf $HOME_DIR/go/src/github.com/intelsdi-x/swan $HOME_DIR
chown -R $VAGRANT_USER:$VAGRANT_USER $HOME_DIR
chown -R $VAGRANT_USER:$VAGRANT_USER /opt/swan

echo "--------------------------- Provisioning done (`date`)"
