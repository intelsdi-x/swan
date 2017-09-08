#!/bin/bash

# Copyright (c) 2017 Intel Corporation
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
# 
#   http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

if [ "$EUID" -ne "0" ]
	then echo "Please run this setup as root"
	return
fi


### Install required packages
echo "Installing required packages..."

# git
apt install git -y

# golang

VERSION="1.9"
PACKAGE="go$VERSION.linux-amd64.tar.gz"

echo "Downloading $PACKAGE ..."
wget https://storage.googleapis.com/golang/$PACKAGE -O /tmp/go.tar.gz
if [ $? -ne 0 ]; then
    	echo "Download failed! Exiting."
        exit 1
fi
echo "Extracting go.tar.gz"
tar -C "$HOME" -xzf /tmp/go.tar.gz

# Be sure to uninstall old go
rm -rf "$HOME/.go"

# Save Go to .go
mv "$HOME/go" "$HOME/.go"
touch "$HOME/.bashrc"
{
	echo 'export GOROOT=$HOME/.go'
	echo 'export PATH=$PATH:$GOROOT/bin'
	
	# Workaround for Glide "Handling default GOPATH for Go 1.8"
	# PR: https://github.com/Masterminds/glide/pull/798
	echo 'export GOPATH=$HOME/go'
	echo 'export PATH=$PATH:$GOPATH/bin'
} >> "$HOME/.bashrc"

export GOROOT=$HOME/.go
export PATH=$PATH:$GOROOT/bin

# Workaround for Glide "Handling default GOPATH for Go 1.8"
# PR: https://github.com/Masterminds/glide/pull/798
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

mkdir -p $HOME/go/{src,pkg,bin}
echo -e "\nGo $VERSION was installed.\n"
rm -f /tmp/go.tar.gz

### Get latest Swan repository
go get github.com/intelsdi-x/swan

### Build Swan
cd ~/go/src/github.com/intelsdi-x/swan/
make build_and_test_unit
cd -
