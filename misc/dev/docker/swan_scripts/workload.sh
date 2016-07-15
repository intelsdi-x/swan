#!/bin/bash

function buildWorkloads() {
    printStep "Build Workloads"
    make build_workloads
    printInfo "Building has been completed."
}

function workload() {
    printStep "Running workload: $SCENARIO"
    cd swan
    buildWorkloads
    BIN=""
    WD=""
    case $SCENARIO in
        "mutilate")
            WD="./workloads/data_caching/memcached/mutilate"
            BIN="./mutilate $BINPARAMETERS"
            ;;
        "memcached")
            WD="./workloads/data_caching/memcached/memcached-1.4.25"
            BIN="./build/memcached -u memcached $BINPARAMETERS"
            ;;
        "caffe")
            WD="./workloads/deep_learning/caffe/caffe_src"
            BIN="./build/tools/caffe $BINPARAMETERS"
            ;;
        "l1d")
            WD="./workloads/low-level-aggressors/"
            BIN="./l1d"
            ;;
        "l1i")
            WD="./workloads/low-level-aggressors/"
            BIN="./l1i"
            ;;
        "l3")
            WD="./workloads/low-level-aggressors/"
            BIN="./l3"
            ;;
        "membw")
            WD="./workloads/low-level-aggressors/"
            BIN="./memBw"
            ;;
        *)
            echo "You must provide scenario for 'workload' target"
            usage
            exit
            ;;
    esac
    printInfo "Executing $BIN from $WD"
    cd $WD
    $BIN
    lockWorkload
}

function lockWorkload() {
    if [[ ! -f /lock.lock ]] && [[ $LOCKSTATE = true ]]; then
        printInfo "Locking container for further usage"
        touch /lock.lock
        sleep inf
    fi

}
