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

if [ -z "$HOME_DIR" ]; then
    HOME_DIR="/home/vagrant"
fi
if [ -z "$VAGRANT_USER" ]; then
    VAGRANT_USER="vagrant"
fi

HOME_DIR=$HOME_DIR SWAN_USER=$VAGRANT_USER $HOME_DIR/go/src/github.com/intelsdi-x/swan/vagrant/provision_experiment_environment.sh

if [ "$SWAN_DEVELOPMENT_ENVIRONMENT" == "true" ]; then
    HOME_DIR=$HOME_DIR SWAN_USER=$VAGRANT_USER $HOME_DIR/go/src/github.com/intelsdi-x/swan/vagrant/provision_development_environment.sh
    HOME_DIR=$HOME_DIR SWAN_USER=$VAGRANT_USER $HOME_DIR/go/src/github.com/intelsdi-x/swan/vagrant/provision_ci_environment.sh
fi
