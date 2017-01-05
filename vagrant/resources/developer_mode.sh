#!/bin/bash

export VAGRANT_USER=vagrant
export HOME_DIR=/home/vagrant

cd /vagrant/resources/scripts

clear
echo "VAGRANT DEVELOPER MODE"
echo "Available keys:"
ssh-add -l

bash
