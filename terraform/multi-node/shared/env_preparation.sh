#!/bin/bash

sudo mkdir -p /vagrant/resources
sudo mkdir -p /cache

sudo chown -R ${USER}:${USER} /vagrant/
mkdir -p ${HOME}/swan_s3_creds/
