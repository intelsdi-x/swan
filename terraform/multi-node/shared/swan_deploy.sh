#!/bin/bash

# This script is desired to prepare enviroment for running swan experiments.
# It's using vagrant scripts for deployment.

set -e

echo "Turn off selinux on remote hosts"
sudo setenforce 0

echo "Setup required enviromental varialbes"
export S3_CREDS_LOCATION="${HOME}/swan_s3_creds/.s3cfg"
export BUCKET_NAME="swan-artifacts"

export HOME_DIR=/home/${USER}
export LD_LIBRARY_PATH=/usr/lib:$LD_LIBRARY_PATH
export VAGRANT_USER=${USER}

export KUBERNETES_VERSION=v1.5.1

pushd /vagrant/resources/scripts

echo "Configuring OS"
sudo -E bash ./copy_configuration.sh
sudo -E bash ./setup_env.sh
sudo -E bash ./install_packages.sh
sudo -E bash ./setup_services.sh
sudo -E bash ./setup_git.sh
sudo -E bash ./post_install.sh

sudo pip install s3cmd==1.6.1

# Depends on image, memcached could be installed by default or not.
sudo adduser memcached || true
sudo systemctl start snapteld

sudo cp -r ${HOME_DIR}/.ssh/* /root/.ssh

echo "Installing kubernetes"
curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBERNETES_VERSION}/kubernetes-server-linux-amd64.tar.gz -o kubernetes-server-linux-amd64.tar.gz

tar xzf kubernetes-server-linux-amd64.tar.gz

sudo cp kubernetes/server/bin/kubectl /usr/bin/kubectl
sudo cp kubernetes/server/bin/kube-apiserver /usr/bin/kube-apiserver
sudo cp kubernetes/server/bin/kube-controller-manager usr/bin/kube-controller-manager
sudo cp kubernetes/server/bin/kube-proxy /usr/bin/kube-proxy
sudo cp kubernetes/server/bin/kube-scheduler /usr/bin/kube-scheduler
sudo cp kubernetes/server/bin/kubelet /usr/bin/kubelet

echo "Download & install swan artifacts"
bash ./artifacts.sh download
sudo bash ./artifacts.sh install

echo "Pull docker image"
s3cmd -c $S3_CREDS_LOCATION sync s3://swan-artifacts/gcr/athena.json ./athena.json
# GCE is skipping email validation and propose using 123@456.com to fulfill docker login requirements.
# For further information: https://cloud.google.com/container-registry/docs/advanced-authentication
sudo docker login -e 123@456.com  -u _json_key -p "$(cat ./athena.json)"  https://gcr.io
sudo docker pull gcr.io/athena-147520/centos_swan_image
sudo docker tag gcr.io/athena-147520/centos_swan_image centos_swan_image

popd
