#!/bin/sh


export SWAN_LOAD_GENERATOR_ADDR=10.5.3.9
#export SWAN_L1D_PATH=/home/ppalucki/golang/src/github.com/intelsdi-x/swan/workloads/low-level-aggressors/l1d
export SWAN_CASSANDRA_ADDR=10.4.3.2
# replace with effect of ansiable deploying mutilate
export SWAN_MEMCACHED_PATH=/home/ppalucki/gopath/src/github.com/intelsdi-x/swan/workloads/data_caching/memcached/memcached-1.4.25/build/memcached

export SWAN_MUTILATE_PATH=/root/sce362/mutilate
export SWAN_MUTILATE_MASTER=10.4.3.1
export SWAN_SLO=500
export SWAN_LOAD_POINTS=2
export SWAN_REPS=1
export SWAN_PEAK_LOAD=1000000
#export SWAN_MUTILATE_AGENT=
export SWAN_SNAPD_ADDR=10.4.3.2
export SWAN_SNAPD_ADDR="none"
# go get github.com/intelsdi-x/snap-plugin-publisher-cassandra
export SWAN_SNAP_CASSANDRA_PLUGIN_PATH=/home/ppalucki/gopath/bin/snap-plugin-publisher-cassandra


pkill memcached
pkill -e -9 mutilate
pssh -P -H 10.4.3.3 -H 10.4.3.4 -H 10.4.3.5 -H 10.4.3.6 -H 10.4.3.7 -H 10.4.3.8 -H 10.4.3.1 pkill -e -9 mutilate

#pstree
#env | grep SWAN_
#echo ready to run ... press a key
#read

echo --- go ---
export GOPATH=/home/ppalucki/gopath
go run /home/ppalucki/gopath/src/github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/*.go --lg_agent=10.4.3.3 --lg_agent=10.4.3.4 --lg_agent=10.4.3.5 --lg_agent=10.4.3.6 --lg_agent=10.4.3.7 --lg_agent=10.4.3.8  --log debug
#/home/ppalucki/gopath/src/github.com/intelsdi-x/swan/build/experiments/memcached/memcached-sensitivity-profile --lg_agent=10.4.3.3 --lg_agent=10.4.3.4 --lg_agent=10.4.3.5 --lg_agent=10.4.3.6 --lg_agent=10.4.3.7 --lg_agent=10.4.3.8 
echo --- done ---
