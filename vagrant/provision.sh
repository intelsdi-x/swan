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

echo `date` "-------------------------- Setting up environment..."
function addEnv() {
    grep "$1" $HOME_DIR/.bash_profile || echo "$1" >> $HOME_DIR/.bash_profile
}
addEnv "export GOPATH=\"$HOME_DIR/go\""
# jupyter intergration tests from notebooks
addEnv 'export PYTHONPATH=$PYTHONPATH:$GOPATH/src/github.com/intelsdi-x/swan'
addEnv 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin'

# Create convenient symlinks in the home directory
ln -sf $HOME_DIR/go/src/github.com/intelsdi-x/swan $HOME_DIR
mkdir -p /opt/swan/bin

echo `date` "Install cassandra configs..."
cp /vagrant/cassandra/cassandra.service /etc/systemd/system

echo `date` "------------------------ Installing centos packages..."
echo `date` "makecache..."
yum makecache fast -y -q

# Warning: takes about 5 minutes
#echo `date` "Update all"
#yum update -y -q

echo `date` "EPEL repo"
yum install -y -q epel-release  # Enables EPEL repo

echo `date` "SWAN deps"
yum install -y -q \
    curl \
    wget \
    python-pip \
    python-devel \
    etcd \
    libcgroup-tools \
    java-1.8.0-openjdk-devel \
    nmap-ncat \
    git \
    sudo

echo `date` "developer tools"
yum install -y -q \
    vim \
    tmux \
    htop


echo `date` "------------------------ Services starting ..."
function daemonStatus() {
    echo "$1 service status: $(systemctl show -p SubState $1 | cut -d'=' -f2)"
}

echo `date` "Reloading systemd..."
systemctl daemon-reload

echo `date` "Docker"
echo `date` "Install docker..."
# https://docs.docker.com/engine/installation/linux/centos/#install-using-the-repository
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum makecache fast -y -q
yum install -y docker-ce
echo `date` "Restart docker..."
systemctl enable docker
systemctl start docker
docker run hello-world
gpasswd -a $VAGRANT_USER docker
daemonStatus docker

echo `date` "ETCD"
systemctl enable etcd
systemctl restart etcd
daemonStatus etcd

echo `date` "Cassandra"
mkdir -p /var/data/cassandra
chcon -Rt svirt_sandbox_file_t /var/data/cassandra # SELinux policy
systemctl enable cassandra
# WARNING takes time - needs to pull cassandra image
echo `date` "Restarting cassandra..."
systemctl restart cassandra
daemonStatus cassandra

# ----------------------------- install snap
# SNAP
echo `date` "Installing snap-telemetry"
curl -s https://packagecloud.io/install/repositories/intelsdi-x/snap/script.rpm.sh | bash
yum install -y snap-telemetry
systemctl enable snap-telemetry
systemctl start snap-telemetry
daemonStatus snap-telemetry

# ----------------------------- install snap plugins
echo `date` "Installing external snap plugins"
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-collector-docker/releases/download/${SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION}/snap-plugin-collector-docker_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-collector-docker
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-publisher-cassandra/releases/download/${SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION}/snap-plugin-publisher-cassandra_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-publisher-cassandra
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-processor-tag/releases/download/${SNAP_PLUGIN_PROCESSOR_TAG_VERSION}/snap-plugin-processor-tag_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-processor-tag
wget --no-verbose https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/${SNAP_PLUGIN_PUBLISHER_FILE_VERSION}/snap-plugin-publisher-file_linux_x86_64 -O ${SWAN_BIN}/snap-plugin-publisher-file

chmod +x ${SWAN_BIN}/snap-plugin-collector-docker
chmod +x ${SWAN_BIN}/snap-plugin-publisher-cassandra
chmod +x ${SWAN_BIN}/snap-plugin-processor-tag
chmod +x ${SWAN_BIN}/snap-plugin-publisher-file

echo `date` "-------------------------- Hyperkube (kubernetes)"
# instead of downloading multiple binaries only hyperkube is downloaded
wget --no-verbose https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/hyperkube -O ${SWAN_BIN}/hyperkube
chmod +x ${SWAN_BIN}/hyperkube

pushd ${SWAN_BIN}
    ./hyperkube --make-symlinks 
popd


echo `date` "--------------------------- Preparing SSH access for root"

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

echo `date` "--------------------------- Go language"
echo `date` "Go downloading..."

GOTGZ=/tmp/go${GO_VERSION}.linux-amd64.tar.gz
wget --no-verbose https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz -O $GOTGZ

echo `date` "Go installing..."
mkdir -p /usr/local
tar -C /usr/local -xzf $GOTGZ 

##################################
# PRIVATE PART
#################################
# -------------------------------- require s3 authoirzation

# # Vagrant user
# touch $HOME_DIR/.ssh/known_hosts
# ssh-keyscan github.com >> $HOME_DIR/.ssh/known_hosts
### Enable this if you require access to private repos.
## ROOT: configure
#git config --global url."git@github.com:".insteadOf "https://github.com/"
# VAGRANT: git rewrite
#executeAsVagrantUser git config --global url."git@github.com:".insteadOf "https://github.com/"
## SSH-agent veryfication
#ssh-add -l

# -------------------------- public keys
if [ -e "$HOME_DIR/swan_s3_creds/.s3cfg" ]; then
    echo `date` "--------------------------- Private components"
    echo `date` "S3 synchronize"
    cp $HOME_DIR/swan_s3_creds/.s3cfg ~/.s3cfg

    echo `date` "Installing s3cmd"
    pip install s3cmd

    echo `date` "Installing public keys"
    s3cmd get s3://swan-artifacts/public_keys authorized_keys
    cat authorized_keys >> ${HOME_DIR}/.ssh/authorized_keys
    
    echo `date` "glide cache"
    s3cmd get s3://swan-artifacts/glide-cache-2017-03-10.tgz /tmp/glide-cache.tgz
    tar --strip-components 2 -C ${HOME_DIR} -xzf /tmp/glide-cache.tgz
    chown -R ${VAGRANT_USER}:${VAGRANT_USER} ${HOME_DIR}/.glide

    echo `date` "sync /opt/swan"
    ### sync everything
    # installing manually requires first to
    # sudo sh -c "mkdir -p /opt/swan && chown -R $USER:$USER /opt/swan"
    s3cmd sync s3://swan-artifacts/workloads/swan/ /opt/swan/

    echo `date` "docker image download"
    s3cmd sync s3://swan-artifacts/workloads/centos_swan_image.tgz /tmp/centos_swan_image.tgz
    echo `date` "docker image import"
    gunzip /tmp/centos_swan_image.tgz 
    docker image import /tmp/centos_swan_image.tar centos_swan_image
fi


# ------------------------- post install
echo `date` "Post install cleaning ...."
# Rewriting permissions
chown -R $VAGRANT_USER:$VAGRANT_USER $HOME_DIR
chown -R $VAGRANT_USER:$VAGRANT_USER /opt/swan

echo `date` "Install binaries to PATH..."
### TODO: resolve issue with PATH
# function executeAsVagrantUser() {
#         sudo -i -u $VAGRANT_USER "$@"
# }
#function addGlobalEnv() {
#    grep "$1" /etc/environment || echo "$1" >> /etc/environment
#}
## Make sure that all required packages are also available for remote access. 
#addGlobalEnv  'PATH=/sbin:/bin:/usr/sbin:/usr/bin:/opt/swan/bin'
ln -sv ${SWAN_BIN}/* /bin/

echo `date` "Provisioning done."
