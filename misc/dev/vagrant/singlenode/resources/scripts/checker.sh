#!/bin/bash

# For debugging purposes only!

. $HOME_DIR/.bash_profile

echo "########## Checking environment... ##########"

echo ">> $HOME_DIR/.bashrc content:"
cat $HOME_DIR/.bash_profile
echo "######################################"
echo ">> $HOME_DIR content:"
ls $HOME_DIR
echo "######################################"

echo "########## Checking GIT configuration... ##########"
echo ">> Root .gitconfig:"
cat /root/.gitconfig
echo "######################################"
echo ">> Root .git directory:"
ls /root/.ssh
echo "######################################"
echo ">> $VAGRANT_USER .gitconfig:"
cat $HOME_DIR/.gitconfig
echo "######################################"
echo ">> $VAGRANT_USER .git directory:"
ls $HOME_DIR/.ssh
echo "######################################"

echo "########## Checking services status... ##########"
echo ">> Docker status:"
systemctl status docker
docker ps
echo "######################################"
echo ">> cassandra status:"
systemctl status cassandra
echo "######################################"
echo ">> etcd status:"
systemctl status etcd
echo "######################################"

echo "########## Checking GO version... ##########"
echo ">> GO version:"
go version
echo "######################################"
echo ">> Glide status:"
glide --version
echo "######################################"

echo "########## Checking snap version... ##########"
echo ">> SNAP version:"
snapctl --version
echo "######################################"
