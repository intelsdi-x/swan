#!/bin/bash

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
} >> "$HOME/.bashrc"

export GOROOT=$HOME/.go
export PATH=$PATH:$GOROOT/bin

mkdir -p $HOME/go/{src,pkg,bin}
echo -e "\nGo $VERSION was installed.\n"
rm -f /tmp/go.tar.gz

### Get latest Swan repository
go get github.com/intelsdi-x/swan

### Build Swan
cd ~/go/src/github.com/intelsdi-x/swan/
make build_and_test_unit
cd -
