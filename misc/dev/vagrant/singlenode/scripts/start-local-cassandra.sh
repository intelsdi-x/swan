#!/bin/bash

set -o nounset -o pipefail -o errexit

# Run this script from inside the development virtual machine
# to provide a local instance of Cassandra for the integration
# test suite.

CASSANDRA_DOCKER_TAG="3.5"
CASSANDRA_HOST_DATA_DIR="/var/data/cassandra"
CASSANDRA_NAME="cassandra-swan"

# Check whether Cassandra is already running

if [[ $(docker ps --no-trunc) == *"$CASSANDRA_NAME"* ]]
then
  echo "Cassandra is already running, nothing to do"
  exit 0
fi

# Start Cassandra

if [[ $(docker ps -a --no-trunc) == *"$CASSANDRA_NAME"* ]]
then
  echo "Removing old container with name $CASSANDRA_NAME"
  docker rm $CASSANDRA_NAME
fi

docker run \
  --name $CASSANDRA_NAME \
  --net host \
  -e CASSANDRA_LISTEN_ADDRESS="127.0.0.1" \
  -e CASSANDRA_CLUSTER_NAME=$CASSANDRA_NAME \
  -v $CASSANDRA_HOST_DATA_DIR:/var/lib/cassandra \
  -d cassandra:$CASSANDRA_DOCKER_TAG

if [[ $(docker ps --no-trunc) == *" $CASSANDRA_NAME "* ]]
then
  echo "Failed to start Cassandra."
  echo "Try checking the logs: 'docker logs $CASSANDRA_NAME'"
  echo "Verify the data directory exists: $CASSANDRA_HOST_DATA_DIR"
  exit 1
fi

# Create the 'snap' keyspace, without which the integration tests fail.

KS="snap"

echo "Creating Cassandra keyspace $KS"

CQLFILE=$(mktemp)
echo "CREATE KEYSPACE IF NOT EXISTS $KS WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };" > $CQLFILE
echo "DESCRIBE KEYSPACES;" >> $CQLFILE

# NB: xargs to trim leading and trailing whitespace
KEYSPACES=$(docker run -it \
  --rm \
  --net host \
  -v /tmp:/tmp \
  cassandra:3.5 \
  cqlsh \
  localhost \
  --file $CQLFILE \
  | xargs)

echo "Keyspaces are: $KEYSPACES"

if [[ $KEYSPACES != *" $KS "* ]]
then
  echo "Creating keyspace '$KS' failed!"
  exit 1;
fi
