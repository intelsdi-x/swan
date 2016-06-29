#!/bin/sh

export SWAN_LOAD_GENERATOR_ADDR=10.5.3.9
#export SWAN_L1D_PATH=/home/ppalucki/golang/src/github.com/intelsdi-x/swan/workloads/low-level-aggressors/l1d
export SWAN_CASSANDRA_ADDR=10.4.3.2
# replace with effect of ansiable deploying mutilate
export SWAN_MEMCACHED_PATH=/home/ppalucki/gopath/src/github.com/intelsdi-x/swan/workloads/data_caching/memcached/memcached-1.4.25/build/memcached

export SWAN_MUTILATE_PATH=/root/sce362/mutilate
export SWAN_MUTILATE_MASTER=10.4.3.1
export SWAN_PEAKLOAD=1000000
export SWAN_MUTILATE_AGENT=

pkill memcached
pkill -9 mutilate
pstree
env | grep SWAN_

echo ready to run ... press a key
read

echo --- go ---
/home/ppalucki/gopath/src/github.com/intelsdi-x/swan/build/experiments/memcached/memcached-sensitivity-profile --mutilate_agent=10.4.3.3 --mutilate_agent=10.4.3.4 --mutilate_agent=10.4.3.5 --mutilate_agent=10.4.3.6 --mutilate_agent=10.4.3.7 --mutilate_agent=10.4.3.8 
echo --- done ---
