#!/bin/bash

set -e

function daemonStatus() {
    echo "$1 service status: $(systemctl show -p SubState $1 | cut -d'=' -f2)"
}

echo "Copying Docker service unit file..."
# docker.service file should be added after docker installation.
cp /vagrant/resources/configs/docker.service /lib/systemd/system/

echo "Reloading systemd..."
systemctl daemon-reload

echo "Configuring docker..."
gpasswd -a $VAGRANT_USER docker
systemctl enable docker
systemctl restart docker
daemonStatus docker

echo "Configuring etcd..."
systemctl enable etcd
systemctl restart etcd
daemonStatus etcd

echo "Configuring cassandra..."
mkdir -p /var/data/cassandra
chcon -Rt svirt_sandbox_file_t /var/data/cassandra # SELinux policy
systemctl enable cassandra.service
systemctl restart cassandra.service 
daemonStatus cassandra
