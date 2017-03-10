echo `date` "Provisioning starting..." 

set -x -e -o pipefail

GO_VERSION="1.7.5"
K8S_VERSION="v1.5.1"
SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION=5
SNAP_PLUGIN_PROCESSOR_TAG_VERSION=3
SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION=5
SNAP_PLUGIN_PUBLISHER_FILE_VERSION=2

# use some sane defaults if it run manually (please run as root)
HOME_DIR="${HOME_DIR:-/home/vagrant}"
VAGRANT_USER="${VAGRANT_USER:-vagrant}"
SWAN_BIN=/opt/swan/bin

# ----------------------- setup env 
echo `date` "Setting up environment..."

# function executeAsVagrantUser() {
#         sudo -i -u $VAGRANT_USER "$@"
# }

function addEnv() {
    grep "$1" $HOME_DIR/.bash_profile || echo "$1" >> $HOME_DIR/.bash_profile
}
## Setting up envs
addEnv "export GOPATH=\"$HOME_DIR/go\""
# addEnv 'export CCACHE_CONFIGPATH=/etc/ccache.conf'
# jupyter intergration tests from notebooks
addEnv 'export PYTHONPATH=$PYTHONPATH:$GOPATH/src/github.com/intelsdi-x/swan'
addEnv 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin'

## Create convenient symlinks in the home directory
ln -sf $HOME_DIR/go/src/github.com/intelsdi-x/swan $HOME_DIR

### TODO: resolve issue with PATH
#function addGlobalEnv() {
#    grep "$1" /etc/environment || echo "$1" >> /etc/environment
#}
## Make sure that all required packages are also available for remote access. 
#addGlobalEnv  'PATH=/sbin:/bin:/usr/sbin:/usr/bin:/opt/swan/bin'

# -------------------- configs
echo `date` "Copying configs..."

mkdir -p /opt/swan/resources
mkdir -p ${SWAN_BIN}

#------------- docker repo
cp /vagrant/resources/configs/docker.repo /etc/yum.repos.d/docker.repo

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

# takes about 5 minutes
#echo `date` "Update all"
#yum update -y -q

echo `date` "EPEL repo"
yum install -y -q epel-release  # Enables EPEL repo

echo `date` "SWAN deps"
yum install -y -q \
    curl \
    wget \
    docker-engine \
    python-pip \
    python-devel \
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

function daemonStatus() {
    echo "$1 service status: $(systemctl show -p SubState $1 | cut -d'=' -f2)"
}

echo `date` "Reloading systemd..."
systemctl daemon-reload

# https://docs.docker.com/engine/installation/linux/centos/#install-using-the-repository
echo `date` "Install docker..."
#yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
#yum makecache fast
#yum install docker-ce
systemctl start docker
docker run hello-world

daemonStatus docker
systemctl enable docker
gpasswd -a $VAGRANT_USER docker

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

# ----------------------------- install snap
# SNAP
echo `date` "Installing snap-telemetry"
curl -s https://packagecloud.io/install/repositories/intelsdi-x/snap/script.rpm.sh | bash
yum install -y snap-telemetry
systemctl enable snap-telemetry
systemctl start snap-telemetry
systemctl status snap-telemetry

echo `date` "Installing external snap plugins"
# PLUGINS

wget --no-verbose https://github.com/intelsdi-x/snap-plugin-collector-docker/releases/download/${SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION}/snap-plugin-collector-docker_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-collector-docker
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-publisher-cassandra/releases/download/${SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION}/snap-plugin-publisher-cassandra_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-publisher-cassandra
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-processor-tag/releases/download/${SNAP_PLUGIN_PROCESSOR_TAG_VERSION}/snap-plugin-processor-tag_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-processor-tag
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/${SNAP_PLUGIN_PUBLISHER_FILE_VERSION}/snap-plugin-publisher-file_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-publisher-file

chmod +x ${SWAN_BIN}/snap-plugin-collector-docker
chmod +x ${SWAN_BIN}/snap-plugin-publisher-cassandra
chmod +x ${SWAN_BIN}/snap-plugin-processor-tag
chmod +x ${SWAN_BIN}/snap-plugin-publisher-file

# -------------------------- KUBERNETEs
echo `date` "Downloading hyperkube"

# instead of downloading multiple binaries only hyperkube is downloaded
wget --no-verbose https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/hyperkube -O ${SWAN_BIN}/hyperkube
chmod +x ${SWAN_BIN}/hyperkube

pushd ${SWAN_BIN}
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

# Generte ssh keys for root - needed to run an experiment
rm -rf /root/.ssh/id_rsa
ssh-keygen -f /root/.ssh/id_rsa -t rsa -N ''
cat /root/.ssh/id_rsa.pub >> /root/.ssh/authorized_keys
chmod og-wx /root/.ssh/authorized_keys

# -------------------------- golang
echo `date` "Downloading golang"

GOTGZ=/tmp/go${GO_VERSION}.linux-amd64.tar.gz
wget --no-verbose https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz -O $GOTGZ

echo `date` "Installing golang"
mkdir -p /usr/local
tar -C /usr/local -xzf $GOTGZ 


##################################
# PRIVATE PART
#################################
# -------------------------------- require s3 authoirzation

### Enable this if you require access to private repos.
## ROOT: configure
#git config --global url."git@github.com:".insteadOf "https://github.com/"
# VAGRANT: git rewrite
#executeAsVagrantUser git config --global url."git@github.com:".insteadOf "https://github.com/"
## SSH-agent veryfication
#ssh-add -l

# -------------------------- public keys
echo `date` "PRIVATE componentes...."
if [ -e "$HOME_DIR/swan_s3_creds/.s3cfg" ]; then
    cp $HOME_DIR/swan_s3_creds/.s3cfg ~/.s3cfg

    echo `date` "Installing s3cmd"
    pip install s3cmd

    echo `date` "Installing public keys"
    s3cmd get s3://swan-artifacts/public_keys authorized_keys
    cat authorized_keys >> ${HOME_DIR}/.ssh/authorized_keys

    # ------------------------- grab all the binaries 
    # low level aggressors from iBench
    s3cmd sync s3://swan-artifacts/workloads/l1d ${SWAN_BIN}/
    s3cmd sync s3://swan-artifacts/workloads/l1i ${SWAN_BIN}/
    s3cmd sync s3://swan-artifacts/workloads/l3 ${SWAN_BIN}/
    s3cmd sync s3://swan-artifacts/workloads/memBw ${SWAN_BIN}/
    # stream 
    s3cmd sync s3://swan-artifacts/workloads/stream.100M ${SWAN_BIN}/

    # memcached
    s3cmd sync s3://swan-artifacts/workloads/mutilate ${SWAN_BIN}/
    s3cmd sync s3://swan-artifacts/workloads/memcached ${SWAN_BIN}/

    # specjbb 
    s3cmd sync s3://swan-artifacts/workloads/specjbb /opt/swan/share/specjbb/

    # caffe
    s3cmd sync s3://swan-artifacts/workloads/caffe /opt/swan/share/caffe/

    # docker image
    s3cmd sync s3://swan-artifacts/workloads/centos_swan_image.tgz /tmp/centos_swan_image.tgz
    gunzip /tmp/centos_swan_image.tgz 
    docker image import /tmp/centos_swan_image.tar centos_swan_image
fi


# ------------------------- post install
echo `date` "Rewriting permissions..."
chown -R $VAGRANT_USER:$VAGRANT_USER $HOME_DIR
chown -R $VAGRANT_USER:$VAGRANT_USER /opt/swan
ln -sv ${SWAN_BIN}/* /bin/


# --------------------------------- as swan user 
#echo `date` "make deps"
#pushd $HOME_DIR/go/src/github.com/intelsdi-x/swan/
#   executeAsVagrantUser make deps_all
#popd

# --------------------------------- make dist && install
# echo "make dist & make install"

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
