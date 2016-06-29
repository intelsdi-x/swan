#!/bin/sh

#export SWAN_L1D_PATH=/home/ppalucki/golang/src/github.com/intelsdi-x/swan/workloads/low-level-aggressors/l1d
# replace with effect of ansiable deploying mutilate


## memcached configuration
export SWAN_MEMCACHED_PATH=/home/ppalucki/gopath/src/github.com/intelsdi-x/swan/workloads/data_caching/memcached/memcached-1.4.25/build/memcached
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
export SWAN_SNAPD_ADDR=10.4.3.2
#export SWAN_SNAPD_ADDR="none"

## cassandra configuration
export SWAN_CASSANDRA_ADDR=10.4.3.2
#go get github.com/intelsdi-x/snap-plugin-publisher-cassandra
export SWAN_SNAP_CASSANDRA_PLUGIN_PATH=/home/ppalucki/gopath/bin/snap-plugin-publisher-cassandra

## clean environment after failure 
snapctl -u http://10.4.3.2:8181 plugin unload collector:mutilate:1
pkill memcached
pkill -e -9 mutilate
pssh -P -H 10.4.3.3 -H 10.4.3.4 -H 10.4.3.5 -H 10.4.3.6 -H 10.4.3.7 -H 10.4.3.8 -H 10.4.3.1 pkill -e -9 mutilate

#pstree
#env | grep SWAN_
#echo ready to run ... press a key
#read

OPTS="--mutilate_agent=10.4.3.3 --mutilate_agent=10.4.3.4 --mutilate_agent=10.4.3.5 --mutilate_agent=10.4.3.6 --mutilate_agent=10.4.3.7 --mutilate_agent=10.4.3.8  --log info"

echo --- go ---
export GOPATH=/home/ppalucki/gopath
EXP_ID=$(go run /home/ppalucki/gopath/src/github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/*.go $OPTS)
#/home/ppalucki/gopath/src/github.com/intelsdi-x/swan/build/experiments/memcached/memcached-sensitivity-profile $OPTS
echo --- done ---

echo --- check logs ---
echo "find /tmp/memcached-sensitivity-profile/*_$EXP_ID/ -type f -exec cat {} \;"
