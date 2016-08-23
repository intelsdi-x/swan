#!/bin/sh
set -x -e

######################## 
# SWAN PART
########################

export SWAN_RUN_ON_KUBERNETES=true
export SWAN_LOG=info

export SWAN_KUBE_LOGLEVEL=4
export SWAN_KUBE_ALLOW_PRIVILEGED=true
export SWAN_KUBE_=true
export SWAN_KUBE_ETCD_SERVERS=http://127.0.0.1:2379

export SWAN_LOAD_POINTS=1
export SWAN_LOAD_DURATION=1s
export SWAN_REPS=1
export SWAN_STOP=true
export SWAN_PEAK_LOAD=10000
export SWAN_AGGR=l3d
export SWAN_HP_COUNT=1

### --------------------------- etcd check and clean
etcdctl --endpoints $SWAN_KUBE_ETCD_SERVERS ls
# clean etcdctl (default registry key)
etcdctl rm --dir --recursive /registry | true

### ---------------------------- docker is running
systemctl status docker

### -------------------------- docker image "centos_swan_image"
# how build docker image
echo sudo docker build -t centos_swan_image -f ./misc/dev/docker/Dockerfile_centos .
sudo docker images centos_swan_image | grep centos_swan_image 

### ------------------------- clean garbage (processes) from previous run
# kill all the remaings from previous unsucessful run
sudo pkill -e kube | true && sleep 1 && ! pgrep kube 
## clean environment after failure
for p in l1d l1i l3 memBw stream.100M memcached mutilate
do
	echo kill $p
	pkill -e -9 $p || true
done

### -------------------------- memcache port is free 
# memcached port is free
! sudo ss -tpln | grep 11211 


### ------------------------- GOPATH - swanpath trick 
# no need to specify path for every binary 
# make the 
sudo ln -fs $GOPATH /opt/gopath 
export GOPATH=/opt/gopath

### --------------------------- snap v0.14 (restart, check) 
# get
go get github.com/intelsdi-x/snap
(cd $GOPATH/src/github.com/intelsdi-x/snap; git checkout v0.14.0-beta)
# build
go install github.com/intelsdi-x/snap
go install github.com/intelsdi-x/snap/cmd/snapctl
# install as service
sudo systemctl stop snap | true
! systemctl status snap
sudo systemd-run --unit snap $GOPATH/bin/snap -t 0 --log-level 1 || true
systemctl status snap

## ----------------------------- swan plugins: mutilate
snapctl plugin load `which snap-plugin-collector-mutilate` || true
snapctl plugin list | grep mutilate
snapctl metric list | grep /intel/swan/mutilate
snapctl plugin unload collector:mutilate:1

## ----------------------------- swan plugins: cassandra
go get github.com/intelsdi-x/snap-plugin-publisher-cassandra
snapctl plugin load `which snap-plugin-publisher-cassandra` | true
snapctl plugin list | grep cassandra

### ------------------------- check mutilate is build
ls -l /opt/gopath/src/github.com/intelsdi-x/swan/workloads/data_caching/memcached/mutilate/mutilate

### ------------------------ run experiment
# compile and run
go install github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile
sudo -E GOPATH=$GOPATH $GOPATH/bin/memcached-sensitivity-profile & 

######################## 
# SERERNITY PART
########################

# ---------------- SLI plugin (memcached)
go get -d github.com/intelsdi-x/serenity2/cmd/serenity
ls -ld $GOPATH/src/github.com/intelsdi-x/serenity2/cmd/serenity
(cd $GOPATH/src/github.com/intelsdi-x/serenity2; git checkout master)
go install github.com/intelsdi-x/serenity2/plugins/snap-collector-memcached
ls -l `which snap-collector-memcached`
snapctl plugin load `which snap-collector-memcached` | true
snapctl plugin list | grep memcached-collector 
snapctl metric list | grep serenity2/memcached

# ---------------- SLO plugin
(cd $GOPATH/src/github.com/intelsdi-x/serenity2; git checkout skonefal/slo_collector)
go install github.com/intelsdi-x/serenity2/plugins/snap-plugin-collector-serenity-slo
ls -l `which snap-plugin-collector-serenity-slo`
snapctl plugin load `which snap-plugin-collector-serenity-slo` | true
snapctl plugin list | grep serenity-slo
snapctl metric list | grep serenity2/lc

# ----------------- FILE publisher
go get github.com/intelsdi-x/snap-plugin-publisher-file
ls -l `which snap-plugin-publisher-file`
snapctl plugin load `which snap-plugin-publisher-file` | true
snapctl plugin list | grep file
# -------------------- SERENITY: run a task sli/slo in background
snapctl task create -t serenity_task.yaml
snapctl task list
ls -l /tmp/serenity-metrics.log

wait

### clean garbabe
# sudo rm -rf local_kube* | true
# sudo rm -rf remote_ulimit* | true

