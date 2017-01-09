#!/bin/bash

set -e

SNAP_VERSION="1.0.0"
SNAP_PLUGIN_COLLECTOR_DOCKER_VERSION=5
SNAP_PLUGIN_PROCESSOR_TAG_VERSION=3
SNAP_PLUGIN_PUBLISHER_CASSANDRA_VERSION=5
SNAP_PLUGIN_PUBLISHER_FILE_VERSION=2

. $HOME_DIR/.bash_profile

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

echo "Installing snap-plugin-publisher-file (version $SNAP_PLUGIN_PUBLISHER_FILE_VERSION)..."
if [ ! -f /cache/snap-plugin-processor-file-${SNAP_PLUGIN_PUBLISHER_FILE_VERSION} ]; then
  wget -q https://github.com/intelsdi-x/snap-plugin-publisher-file/releases/download/${SNAP_PLUGIN_PUBLISHER_FILE_VERSION}/snap-plugin-publisher-file_linux_x86_64 -O $GOPATH/bin/snap-plugin-publisher-file
  chmod +x  $GOPATH/bin/snap-plugin-publisher-file
  touch /cache/snap-plugin-publisher-file-${SNAP_PLUGIN_PUBLISHER_FILE_VERSION}
fi