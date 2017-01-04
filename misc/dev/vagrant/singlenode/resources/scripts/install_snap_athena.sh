#!/bin/bash

set -e -x

SNAP_VERSION="1.0.0"
SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION=5
SNAP_PLUGIN_PROCESSOR_TAG_VERSION=3
SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION=5

. $HOME_DIR/.bash_profile
ATHENA_DIR=$GOPATH/src/github.com/intelsdi-x/athena

echo "Installing Snap..."
if [ ! -f /cache/snap-${SNAP_VERSION}-linux-amd64.tar.gz ]; then
    wget -q -P /cache https://github.com/intelsdi-x/snap/releases/download/${SNAP_VERSION}/snap-${SNAP_VERSION}-linux-amd64.tar.gz
    tar xf /cache/snap-${SNAP_VERSION}-linux-amd64.tar.gz -C /cache
    mv /cache/snaptel $GOPATH/bin
    mv /cache/snapteld $GOPATH/bin
fi

echo "Installing snap-plugin-collector-docker (version $SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION)..."
if [ ! -f /cache/snap-plugin-collector-docker-${SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION} ]; then
    wget -q https://github.com/intelsdi-x/snap-plugin-collector-docker/releases/download/${SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION}/snap-plugin-collector-docker_linux_x86_64 -O $GOPATH/bin/snap-plugin-collector-docker
    chmod +x $GOPATH/bin/snap-plugin-collector-docker
    touch /cache/snap-plugin-collector-docker-${SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION}
fi

echo "Installing snap-plugin-publisher-cassandra (version $SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION)..."
if [ ! -f /cache/snap-plugin-publisher-cassandra-${SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION} ]; then
    wget -q https://github.com/intelsdi-x/snap-plugin-publisher-cassandra/releases/download/${SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION}/snap-plugin-publisher-cassandra_linux_x86_64 -O $GOPATH/bin/snap-plugin-publisher-cassandra
    chmod +x  $GOPATH/bin/snap-plugin-publisher-cassandra
    touch /cache/snap-plugin-publisher-cassandra-${SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION}
fi

echo "Installing snap-plugin-processor-tag (version $SNAP_PLUGIN_PROCESSOR_TAG_VERSION)..."
if [ ! -f /cache/snap-plugin-processor-tag-${SNAP_PLUGIN_PROCESSOR_TAG_VERSION} ]; then
  wget -q https://github.com/intelsdi-x/snap-plugin-processor-tag/releases/download/${SNAP_PLUGIN_PROCESSOR_TAG_VERSION}/snap-plugin-processor-tag_linux_x86_64 -O $GOPATH/bin/snap-plugin-processor-tag
  chmod +x  $GOPATH/bin/snap-plugin-processor-tag
  touch /cache/snap-plugin-processor-tag-${SNAP_PLUGIN_PROCESSOR_TAG_VERSION}
fi


echo "Installing Athena & its K8s..."
if [ ! -d $ATHENA_DIR ]; then
    echo "Fetching Athena sources"
    mkdir -p $ATHENA_DIR
    git clone git@github.com:intelsdi-x/athena $ATHENA_DIR
else
    echo "Updating Athena sources"
    pushd $ATHENA_DIR
    git pull
    popd
fi
echo "Fetching kubernetes binaries for Athena"
cd $ATHENA_DIR && ./misc/kubernetes/install_binaries.sh
