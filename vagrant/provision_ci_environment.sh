#!/usr/bin/env bash
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

if [ "$USER" != "root" ]; then
    echo "This script needs to be run with root privileges"
    exit 1
fi
if [ "$SWAN_USER" == "" ]; then
    echo "You need to set SWAN_USER environmental variable"
    exit 1
fi
if [ "$HOME_DIR" == "" ]; then
    echo "You need to set HOME_DIR environmental variable"
    exit 1
fi


echo "-------------------------- CI environment (`date`)"
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


    echo "Downloading private dependencies"
    s3cmd get --recursive s3://swan-artifacts/workloads/swan-private/  /opt/swan/
fi
echo "--------------------------- Provisioning CI environment done (`date`)"

echo "--------------------------- Post install (`date`)"
ln -svf $HOME_DIR/go/src/github.com/intelsdi-x/swan $HOME_DIR
chown -R $SWAN_USER:$SWAN_USER $HOME_DIR
chown -R $SWAN_USER:$SWAN_USER /opt/swan
chmod -R +x /opt/swan/bin/*
ln -svf /opt/swan/bin/* /usr/bin/

echo "--------------------------- Provisioning development environment done (`date`)"
