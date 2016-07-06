#!/bin/bash
set -o pipefail

REPOSITORY_URL=github.com/intelsdi-x/swan

TARGET=""
SCENARIO=""
BINPARAMETERS=""

function printError() {
    echo -e "\033[0;31m### $1 ###\033[0m"
}

function printStep() {
    echo -e "\033[0;36m[$1]\033[0m"
}

function printInfo() {
    echo -e "\033[0;32m> $1\033[0m"
}

function printOption() {
    echo -e " * '-$1'\t$2"
}

function verifyStatus() {
    if [[ !$? -eq 0 ]]; then
        printError "Step doesn't exit with exit code 0"
        exit 1
    fi
}

function setGitHubCredentials() {
    printStep "Set GitHub credentials"
    if [[ $GIT_TOKEN != "" ]]; then
        git config --global url."https://$GIT_TOKEN:x-oauth-basic@github.com/".insteadOf "https://github.com/"
        printInfo "Token for GitHub has been set"
    fi
}

function buildSnap() {
    printStep "Build snap"
    git clone https://github.com/intelsdi-x/snap
    cd snap
    make deps
    make all
    verifyStatus
    printInfo "Snap has been build"
    cd ..
}

function cloneCode() {
    printStep "Clone source code from github"

    if [[ $GIT_BRANCH == "" ]]; then
        GIT_BRANCH="master"
    fi

    printInfo "Selected branch: $GIT_BRANCH"
    git clone -b $GIT_BRANCH  https://$REPOSITORY_URL

    verifyStatus
    printInfo "Clone source code has been completed"
}

function prepareEnvironment() {
    printStep "Prepare environment"
    make deps
    verifyStatus
    printInfo "All dependencies have been downloaded"
}

function runScenarioMake() {
    printStep "Running scenario: $SCENARIO"
    DEFAULT_TARGETS=all
    if [[ "$SCENARIO" != "" ]]; then
        DEFAULT_TARGETS="$@"
    fi

    printInfo "Selected targets: $DEFAULT_TARGETS"

    make $DEFAULT_TARGETS
    verifyStatus
}

function getCodeFromDir() {
    printStep "Binding source code from /swan to proper directory in \$GOPATH"
    mkdir swan
    mount -o bind /swan ./swan
    verifyStatus
    printInfo "Binding has been completed"
}

function getCode() {
    if [[ -d "/swan" ]]; then
        getCodeFromDir
    else
        cloneCode
    fi
}

function runAndPrepareMakeTarget() {
    printStep "Preparing environment for running 'make $SCENARIO'"
    buildSnap
    cd swan
    prepareEnvironment
    runScenarioMake
}

function usage() {
    echo "Usage:"
    printOption "t" "Run selected target. Possible choices are: 'make' and 'workload'. Default: 'make'"
    printOption "s" "Selected scenario for target. Possible choices are: 'make': options are specified in swan's Makefile; default: 'integration_test', 'workload':['caffe', 'memcached', 'mutilate']; default: 'memcached'"
    printOption "p" "Pass parameters to workload binaries. Only for 'workload' target. There is no default parameter."
}

function parseArguments() {
    printStep "Parsing arguments"
    while getopts "t:s:p:" opt; do
    case $opt in
        t)
            TARGET=$OPTARG
            ;;
        s)
            SCENARIO=$OPTARG
            ;;
        p)
            BINPARAMETERS="$OPTARG"
            ;;
        *)
            usage
            exit
            ;;
        esac
    done
}

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
    case $SCENARIO in
        "mutilate")
            BIN="./workloads/data_caching/memcached/mutilate/mutilate $BINPARAMETERS"
            ;;
        "memcached")
            BIN="./workloads/data_caching/memcached/memcached-1.4.25/build/memcached -u memcached $BINPARAMETERS"
            ;;
        "caffe")
            BIN="./workloads/deep_learning/caffe/caffe_src/build/tools/caffe $BINPARAMETERS"
            ;;
        *)
            echo "You must provide scenario for 'workload' target"
            usage
            exit
            ;;
    esac
    printInfo "Executing $BIN"
    $BIN
}

function main() {
    printInfo "Configuring source code repository"
    setGitHubCredentials
    getCode
    parseArguments "$@"
    case $TARGET in
        "workload")
            workload
            ;;
        *)
            runAndPrepareMakeTarget
            ;;
    esac
}

main "$@"
