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
export SWAN_MEMCACHED_THREADS=8
export SWAN_MEMCACHED_CONNECTIONS=16000

## mutilate configuration
export SWAN_MUTILATE_PATH=/root/sce362/mutilate
export SWAN_MUTILATE_MASTER=10.4.3.1
export SWAN_MUTILATE_MASTER_THREADS=20
export SWAN_MUTILATE_MASTER_CONNECTIONS=8
export SWAN_MUTILATE_MASTER_DEPTH=4
export SWAN_MUTILATE_WARMUP_TIME=5s
export SWAN_MUTILATE_AGENT_THREADS=20
export SWAN_MUTILATE_AGENT_CONNECTIONS=8
export SWAN_MUTILATE_AGENT_DEPTH=4
export SWAN_MUTILATE_AGENT_PORT=6556
export SWAN_MUTILATE_AGENT_AFFINITY=true

## experiment configuration
export SWAN_SLO=500
export SWAN_LOAD_DURATION=60s
#export SWAN_PEAK_LOAD=1000000
#export SWAN_LOAD_POINTS=10
export SWAN_PEAK_LOAD=900000
export SWAN_LOAD_POINTS=1
export SWAN_REPS=3
export SWAN_LOG=info

## isolations
#NUMA node0 CPU(s):     0-7,16-23
#NUMA node1 CPU(s):     8-15,24-31
# 0+16 is sibiling pair
export SWAN_HP_SETS=0,1,2,3,4,5,6,7:0
export SWAN_BE_SETS=16,17,18,19,20,21,22,23:0

## snap configuration
export SWAN_SNAPD_ADDR=10.4.3.9
#for disabling snap
#export SWAN_SNAPD_ADDR="none"

#export SWAN_AGGR=stream
export SWAN_AGGR=l1d,l3d,stream

## cassandra configuration
export SWAN_CASSANDRA_ADDR=10.4.3.10

export SWAN_SNAP_CASSANDRA_PLUGIN_PATH=$GOPATH/bin/snap-plugin-publisher-cassandra

if [ ! -f $SWAN_SNAP_CASSANDRA_PLUGIN_PATH ]; then
	go get -v github.com/intelsdi-x/snap-plugin-publisher-cassandra
fi

if false; then
	echo -- run and check cassandra --
	systemd-run -H $SWAN_CASSANDRA_ADDR --unit cassandra_docker -r docker run --name cassandra_docker --net host -e CASSANDRA_LISTEN_ADDRESS=10.4.3.10 -e CASSANDRA_CLUSTER_NAME=casssandra-docker -v /var/data/cassandra:/var/lib/casssandra cassandra:3.5 || true
	systemctl -H $SWAN_CASSANDRA_ADDR status cassandra_docker
	scp /home/ppalucki/gopath/src/github.com/intelsdi-x/swan/misc/dev/vagrant/singlenode/resources/keyspace.cql 10.4.3.10:/root/sce362/
	ssh 10.4.3.10 docker run --rm --net host  -v /root/sce362/keyspace.cql:/resources/keyspace.cql cassandra:3.5 cqlsh localhost --file /resources/keyspace.cql
fi

if false; then
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
fi

if true; then
	echo -- cleaning env ---
	## clean environment after failure 
	pkill memcached || true
	pkill -e -9 mutilate || true
	pssh -P -H 10.4.3.3 -H 10.4.3.4 -H 10.4.3.5 -H 10.4.3.6 -H 10.4.3.7 -H 10.4.3.8 -H 10.4.3.1 pkill -e -9 mutilate || true
	pssh -p 1 -i -H 10.4.3.3 -H 10.4.3.4 -H 10.4.3.5 -H 10.4.3.6 -H 10.4.3.7 -H 10.4.3.8 -H 10.4.3.1 pgrep mutilate || true
fi

# check swan lab now!!!
pstree
#env | grep SWAN_
echo ready to run ... press a key
read


echo -- experiment ---
OPTS="--mutilate_agent=10.4.3.3 --mutilate_agent=10.4.3.4 --mutilate_agent=10.4.3.5 --mutilate_agent=10.4.3.6 --mutilate_agent=10.4.3.7 --mutilate_agent=10.4.3.8"
echo --- go ---
EXP_ID=$(go run $GOPATH/src/github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/*.go $OPTS)
#$GOPATH/src/github.com/intelsdi-x/swan/build/experiments/memcached/memcached-sensitivity-profile $OPTS
echo --- done ---

echo $EXP_ID

echo --- profile ---
$GOPATH/src/github.com/intelsdi-x/swan/build/viewer/sensitivity_viewer sensitivity --cassandra_host="$SWAN_CASSANDRA_ADDR" $EXP_ID

echo --- check logs ---
echo "find /tmp/memcached-sensitivity-profile/*_$EXP_ID/ -type f -exec cat {} \;"
