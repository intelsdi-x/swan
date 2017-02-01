#!/bin/bash

export GOPATH="/usr"
export SWAN_LOG="debug"
export SWAN_SLO="500"
export SWAN_PEAK_LOAD="500000"
export SWAN_REPS="1"
export SWAN_STOP="true"
export SWAN_AGGR="caffe"
export SWAN_LOAD_POINTS="1"
export SWAN_LOAD_DURATION="5s"
# Kubernetes
export SWAN_RUN_ON_KUBERNETES="true"
export SWAN_HP_KUBERNETES_MEMORY_RESOURCE="5000000000"
export SWAN_KUBE_ALLOW_PRIVILEGED="true"
export SWAN_KUBE_APISERVER_ARGS="--insecure-bind-address=${controller_ip}"
export SWAN_KUBE_LOGLEVEL="4"
export SWAN_KUBERNETES_MASTER="${controller_ip}"
# Experiment isolations
export SWAN_HP_SETS="0,1,2,3:0"
export SWAN_BE_SETS="4,5,6,7,12,13,14,15:0"
export SWAN_BE_L1_SETS="8,9,10,11:0"
# Experiment deployment
export SWAN_SNAPD_ADDR="http://${controller_ip}:8181"
export SWAN_CASSANDRA_ADDR="${node1}"
# HP configution
export SWAN_MEMCACHED_THREADS="4"
export SWAN_MEMCACHED_CONNECTIONS="1024"
export SWAN_MEMCACHED_IP="${controller_ip}"
export SWAN_MEMCACHED_THREADS_AFFINITY="true"
export SWAN_MEMCACHED_MAX_MEMORY="4000"
# Aggressors configuration
export SWAN_L1D_PROCESS_NUMBER="4"
export SWAN_L1I_PROCESS_NUMBER="4"
export SWAN_L3_PROCESS_NUMBER="8"
export SWAN_MEMBW_PROCESS_NUMBER="8"
export SWAN_MUTILATE_RECORDS="1000000"
export SWAN_MUTILATE_MASTER_THREADS="8"
export SWAN_MUTILATE_MASTER_CONNECTIONS="4"
export SWAN_MUTILATE_MASTER_CONNECTIONS_DEPTH="4"
export SWAN_MUTILATE_MASTER_AFFINITY="true"
export SWAN_MUTILATE_MASTER_BLOCKING="true"
export SWAN_MUTILATE_AGENT_THREADS="8"
export SWAN_MUTILATE_AGENT_AFFINITY="true"
export SWAN_MUTILATE_AGENT_BLOCKING="true"
export SWAN_MUTILATE_AGENT_CONNECTIONS="25"
export SWAN_MUTILATE_AGENT_CONNECTIONS_DEPTH="8"
export SWAN_MUTILATE_MASTER="${node1}"
export SWAN_MUTILATE_AGENT="${node2},${node3}"

sudo -E memcached-sensitivity-profile --kubelet_path=/usr/bin/kubelet --kube_apiserver_path=/usr/bin/kube-apiserver --kube_controller_path=/usr/bin/kube-controller-manager --kube_proxy_path=/usr/bin/kube-proxy --kube_scheduler_path=/usr/bin/kube-scheduler --snap_plugins_path=/usr/bin
