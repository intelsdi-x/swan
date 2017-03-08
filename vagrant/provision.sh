echo `date` "Provisioning starting..." 

set -x -e -o pipefail
# ----------------------- setup env 
echo `date` "Setting up environment..."
function addEnv() {
    grep "$1" $HOME_DIR/.bash_profile || echo "$1" >> $HOME_DIR/.bash_profile
}
function addGlobalEnv() {
    grep "$1" /etc/environment || echo "$1" >> /etc/environment
}
function executeAsVagrantUser() {
        sudo -E -u $VAGRANT_USER -s PATH=$PATH GOPATH=$GOPATH CCACHECONFDIR=$CCACHECONFDIR "$@"
}
function daemonStatus() {
    echo "$1 service status: $(systemctl show -p SubState $1 | cut -d'=' -f2)"
}

## Setting up envs
addEnv "export GOPATH=\"$HOME_DIR/go\""
addEnv 'export CCACHE_CONFIGPATH=/etc/ccache.conf'
# jupyter intergration tests from notebooks
addEnv 'export PYTHONPATH=$PYTHONPATH:$GOPATH/src/github.com/intelsdi-x/swan'
addEnv 'export PATH=/usr/lib64/ccache/:$PATH:/opt/swan/bin:/usr/local/go/bin:$GOPATH/bin'

## Create convenient symlinks in the home directory
ln -sf $HOME_DIR/go/src/github.com/intelsdi-x/swan $HOME_DIR

## Make sure that all required packages are also available for remote access. 
addGlobalEnv  'PATH=/usr/lib64/ccache:/sbin:/bin:/usr/sbin:/usr/bin:/opt/swan/bin'

# -------------------- configs
echo `date` "Copying configs..."

mkdir -p /opt/swan/resources
mkdir -p /cache/ccache


#------------- docker repo
cp /vagrant/resources/configs/docker.repo /etc/yum.repos.d/docker.repo

# ccache
#cp /vagrant/resources/configs/ccache.conf /etc/ccache.conf

# yum optmizie
#cp /vagrant/resources/configs/fastestmirror.conf /etc/yum/pluginconf.d/fastestmirror.conf

# ----------------------- cassandra service
cp /vagrant/resources/configs/cassandra.service /etc/systemd/system
cp /vagrant/resources/configs/keyspace.cql /opt/swan/resources
cp /vagrant/resources/configs/table.cql /opt/swan/resources


# ------------------------- PACKAGES
echo `date` "Installing centos packages..."

echo `date` "Makecache..."
yum makecache fast -y -q
echo `date` "Update all"
yum update -y -q
echo `date` "EPEL repo"
yum install -y -q epel-release  # Enables EPEL repo
echo `date` "SWAN deps"
yum install -y -q \
    curl \
    wget \
    python-pip \
    docker-engine \
    etcd \
    java-1.8.0-openjdk-devel \
    git \
    sudo

# echo Installing packages
# yum groupinstall -y -q "Development tools"
# yum install -y -q \
#     boost \
#     boost-devel \
#     ccache \
#     cppzmq-devel \
#     deltarpm \
#     docker-engine \
#     etcd \
#     gcc-g++ \
#     gengetopt \
#     gflags \
#     gflags-devel \
#     git \
#     glog \
#     glog-devel \
#     hdf5 \
#     hdf5-devel \
#     hg \
#     htop \
#     iptables \
#     java-1.8.0-openjdk-devel \
#     leveldb \
#     leveldb-devel \
#     libcgroup-tools \
#     libevent-devel \
#     lmdb \
#     lmdb-devel \
#     moreutils-parallel \
#     nmap-ncat \
#     numactl \
#     openblas \
#     openblas-devel \
#     opencv \
#     opencv-devel \
#     perf \
#     protobuf \
#     protobuf-devel \
#     psmisc \
#     pssh \
#     python-pip \
#     python-devel \
#     snappy \
#     snappy-devel \
#     scons \
#     sudo \
#     tree \
#     vim \
#     wget \
#     zeromq-devel
# yum clean all
#

echo `date` "Reloading systemd..."
systemctl daemon-reload

echo `date` "Configuring docker..."
gpasswd -a $VAGRANT_USER docker
systemctl enable docker
systemctl restart docker
daemonStatus docker

echo `date` "Configuring etcd..."
systemctl enable etcd
systemctl restart etcd
daemonStatus etcd

# WARNING pulls docker image
echo `date` "Configuring cassandra..."
mkdir -p /var/data/cassandra
chcon -Rt svirt_sandbox_file_t /var/data/cassandra # SELinux policy
systemctl enable cassandra.service
echo `date` "Restarting cassandra..."
systemctl restart cassandra.service 
daemonStatus cassandra

# ----------------------------------------- WGET/CURL DOWNLOADING
# -------------------------- golang
echo `date` "Installing golang"
GO_VERSION="1.7.3"
mkdir -p /cache
curl -s https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz -O /cache/go${GO_VERSION}.linux-amd64.tar.gz
tar xf /cache/go${GO_VERSION}.linux-amd64.tar.gz -C /usr/local

# ----------------------------- install snap
# SNAP
echo `date` "Installing snap-telemetry"
curl -s https://packagecloud.io/install/repositories/intelsdi-x/snap/script.rpm.sh | sudo bash
yum install -y snap-telemetry
systemctl enable snap-telemetry
systemctl start snap-telemetry
systemctl status snap-telemetry

echo `date` "Installing external snap plugins"
# PLUGINS
SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION=5
SNAP_PLUGIN_PROCESSOR_TAG_VERSION=3
SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION=5
SNAP_PLUGIN_PUBLISHER_FILE_VERSION=2
wget https://github.com/intelsdi-x/snap-plugin-collector-docker/releases/download/${SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION}/snap-plugin-collector-docker_linux_x86_64 -O /opt/swan/bin/snap-plugin-collector-docker
wget https://github.com/intelsdi-x/snap-plugin-publisher-cassandra/releases/download/${SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION}/snap-plugin-publisher-cassandra_linux_x86_64 -O /opt/swan/bin/snap-plugin-publisher-cassandra
wget https://github.com/intelsdi-x/snap-plugin-processor-tag/releases/download/${SNAP_PLUGIN_PROCESSOR_TAG_VERSION}/snap-plugin-processor-tag_linux_x86_64 -O /opt/swan/bin/snap-plugin-processor-tag
wget https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/${SNAP_PLUGIN_PUBLISHER_FILE_VERSION}/snap-plugin-publisher-file_linux_x86_64 -O /opt/swan/bin/snap-plugin-publisher-file

# -------------------------- KUBERNETEs
echo `date` "Downloading hyperkube"
K8S_VERSION="v1.5.1"
# instead of downloading multiple binaries only hyperkube is downloaded
wget -q https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/hyperkube -O /opt/swan/bin/hyperkube
chmod +x /opt/swan/bin/hyperkube

pushd /opt/swan/bin

./hyperkube --make-symlinks 
popd

# ------------------------------- git setup
echo `date` "Preparing SSH environment for root and $VAGRANT_USER"

## Vagrant user
touch $HOME_DIR/.ssh/known_hosts
ssh-keyscan github.com >> $HOME_DIR/.ssh/known_hosts

## ROOT
mkdir -p /root/.ssh

# known hosts
touch /root/.ssh/known_hosts
ssh-keyscan github.com >> /root/.ssh/known_hosts
ssh-keyscan localhost >> /root/.ssh/known_hosts
ssh-keyscan 127.0.0.1 >> /root/.ssh/known_hosts

# Add key to SSH agent (fail when no ssh-agent is accessible, one won't be able to download private repos)
# Add ssh keys for root - needed to run an experiment
rm -rf /root/.ssh/id_rsa
ssh-keygen -f /root/.ssh/id_rsa -t rsa -N ''
cat /root/.ssh/id_rsa.pub >> /root/.ssh/authorized_keys
chmod og-wx /root/.ssh/authorized_keys

## ROOT: configure
git config --global url."git@github.com:".insteadOf "https://github.com/"
# VAGRANT: git rewrite
executeAsVagrantUser git config --global url."git@github.com:".insteadOf "https://github.com/"

## SSH-agent veryfication
ssh-add -l

# -------------------------------- require s3 authoirzation

echo `date` "Installing python packages"
pip install s3cmd

echo `date` "Installing SpecJBB"
pushd $HOME_DIR/go/src/github.com/intelsdi-x/swan/
    ./scripts/get_specjbb.sh -s . -c $HOME_DIR/swan_s3_creds/.s3cfg -b swan-artifacts
popd

# -------------------------- public keys
echo `date` "Installing public keys"
if [ -e "$HOME_DIR/swan_s3_creds/.s3cfg" ]; then
    s3cmd get -c $HOME_DIR/swan_s3_creds/.s3cfg s3://swan-artifacts/public_keys authorized_keys
    cat authorized_keys >> ${HOME_DIR}/.ssh/authorized_keys
fi

# ------------------------- post install
echo `date` "Rewriting permissions..."
chown -R $VAGRANT_USER:$VAGRANT_USER $HOME_DIR
chown -R $VAGRANT_USER:$VAGRANT_USER /cache

# as swan user 

# --------------------------------- make dist && install
# echo "make dist & make install"
#
#
# BUILD_OPENBLAS=""
#
# pushd $HOME_DIR/go/src/github.com/intelsdi-x/swan/
#     
#     executeAsVagrantUser make repository_reset
#     executeAsVagrantUser make deps_all
#     if [[ "$BUILD_DOCKER_IMAGE" == "true" ]]; then
#             executeAsVagrantUser make dist
#             executeAsVagrantUser make build_image
#     else
#             executeAsVagrantUser make dist
#     fi
#
#     make install
# popd
echo `date` "Provisioning done."
