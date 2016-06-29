#!/bin/sh

set -e
set -x

# memcached connection reset by peer fix
echo 0 > /proc/sys/net/ipv4/tcp_syncookies

# all code below depdens on it
export GOPATH=/home/ppalucki/gopath

#export SWAN_L1D_PATH=$GOPATH/src/github.com/intelsdi-x/swan/workloads/low-level-aggressors/l1d

## memcached configuration
export SWAN_MEMCACHED_PATH=$GOPATH/src/github.com/intelsdi-x/swan/workloads/data_caching/memcached/memcached-1.4.25/build/memcached
export SWAN_MEMCACHED_IP=10.5.3.9

## mutilate configuration
export SWAN_MUTILATE_PATH=/root/sce362/mutilate
export SWAN_MUTILATE_MASTER=10.4.3.1
export SWAN_MUTILATE_MASTER_THREADS=1
export SWAN_MUTILATE_MASTER_CONNECTIONS=1
export SWAN_MUTILATE_MASTER_DEPTH=1
export SWAN_MUTILATE_WARMUP_TIME=0
export SWAN_MUTILATE_AGENT_THREADS=1
export SWAN_MUTILATE_AGENT_CONNECTIONS=1
export SWAN_MUTILATE_AGENT_DEPTH=1
export SWAN_MUTILATE_AGENT_PORT=6556

## experiment configuration
export SWAN_SLO=500
export SWAN_LOAD_POINTS=1
export SWAN_LOAD_DURATION=1
export SWAN_REPS=1
export SWAN_PEAK_LOAD=1000000

## snap configuration
export SWAN_SNAPD_ADDR=10.4.3.10
#for disabling snap
#export SWAN_SNAPD_ADDR="none"

## cassandra configuration
export SWAN_CASSANDRA_ADDR=10.4.3.10
go get github.com/intelsdi-x/snap-plugin-publisher-cassandra
export SWAN_SNAP_CASSANDRA_PLUGIN_PATH=$GOPATH/bin/snap-plugin-publisher-cassandra

echo -- run and check cassandra --
systemd-run -H $SWAN_CASSANDRA_ADDR --unit cassandra_docker -r docker run --name cassandra_docker --net host -e CASSANDRA_LISTEN_ADDRESS=10.4.3.10 -e CASSANDRA_CLUSTER_NAME=casssandra-docker -v /var/data/cassandra:/var/lib/casssandra cassandra:3.5 || true
systemctl -H $SWAN_CASSANDRA_ADDR status cassandra_docker
scp /home/ppalucki/gopath/src/github.com/intelsdi-x/swan/misc/dev/vagrant/singlenode/resources/keyspace.cql 10.4.3.10:/root/sce362/
ssh 10.4.3.10 docker run --rm --net host  -v /root/sce362/keyspace.cql:/resources/keyspace.cql cassandra:3.5 cqlsh localhost --file /resources/keyspace.cql


echo -- build, copy, run and check snap --
# get and build
go get -v github.com/intelsdi-x/snap
# cp
scp $GOPATH/bin/snap $SWAN_SNAPD_ADDR:/root/sce362/ || true
# reset previous
systemctl -H $SWAN_SNAPD_ADDR reset-failed
# run
systemd-run -H $SWAN_SNAPD_ADDR --unit=snap /root/sce362/snap -t 0 --log-level 1 || true
# check
systemctl -H $SWAN_SNAPD_ADDR status snap
snapctl -u http://$SWAN_SNAPD_ADDR:8181 plugin load $GOPATH/src/github.com/intelsdi-x/swan/build/snap-plugin-collector-mutilate || true
snapctl -u http://$SWAN_SNAPD_ADDR:8181 plugin load $SWAN_SNAP_CASSANDRA_PLUGIN_PATH || true
snapctl -u http://$SWAN_SNAPD_ADDR:8181 plugin list
snapctl -u http://$SWAN_SNAPD_ADDR:8181 metric list
snapctl -u http://$SWAN_SNAPD_ADDR:8181 plugin unload collector:mutilate:1


echo -- cleaning env ---
## clean environment after failure 
pkill memcached || true
pkill -e -9 mutilate || true
pssh -P -H 10.4.3.3 -H 10.4.3.4 -H 10.4.3.5 -H 10.4.3.6 -H 10.4.3.7 -H 10.4.3.8 -H 10.4.3.1 pkill -e -9 mutilate || true

#pstree
#env | grep SWAN_
#echo ready to run ... press a key
#read


echo -- experiment ---
OPTS="--mutilate_agent=10.4.3.3 --mutilate_agent=10.4.3.4 --mutilate_agent=10.4.3.5 --mutilate_agent=10.4.3.6 --mutilate_agent=10.4.3.7 --mutilate_agent=10.4.3.8  --log info --aggr=l1d"

echo --- go ---
EXP_ID=$(go run $GOPATH/src/github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/*.go $OPTS)
#$GOPATH/src/github.com/intelsdi-x/swan/build/experiments/memcached/memcached-sensitivity-profile $OPTS
echo --- done ---

echo $EXP_ID

echo --- profile ---
$GOPATH/src/github.com/intelsdi-x/swan/build/viewer/sensitivity_viewer sensitivity --cassandra_host="$SWAN_CASSANDRA_ADDR" $EXP_ID

echo --- check logs ---
echo "find /tmp/memcached-sensitivity-profile/*_$EXP_ID/ -type f -exec cat {} \;"
