#!/bin/bash

set -e

echo "Setting up git..."

## Preparing SSH environment for $VAGRANT_USER
touch $HOME_DIR/.ssh/known_hosts
grep github.com $HOME_DIR/.ssh/known_hosts || ssh-keyscan github.com >> $HOME_DIR/.ssh/known_hosts
sudo -u $VAGRANT_USER git config --global url."git@github.com:".insteadOf "https://github.com/"

## Preparing SSH environment for root
mkdir -p ~/.ssh
touch ~/.ssh/known_hosts
grep github.com ~/.ssh/known_hosts || ssh-keyscan github.com >> ~/.ssh/known_hosts
git config --global url."git@github.com:".insteadOf "https://github.com/"

# Add key to SSH agent (fail when no ssh-agent is accessible, one won't be able to download private repos)
# Add ssh keys for root - needed to run an experiment
rm -fr /root/.ssh/id_rsa
ssh-keygen -f /root/.ssh/id_rsa -t rsa -N ''
cat /root/.ssh/id_rsa.pub >> /root/.ssh/authorized_keys
chmod og-wx /root/.ssh/authorized_keys
ssh-keyscan localhost >> /root/.ssh/known_hosts

## SSH-agent veryfication
ssh-add -l
