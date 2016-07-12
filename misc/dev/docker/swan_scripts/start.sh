#!/bin/bash
set -o pipefail

REPOSITORY_URL=github.com/intelsdi-x/swan

TARGET=""
SCENARIO=""
BINPARAMETERS=""
LOCKSTATE=false
COLORTERMINAL=false

. /make.sh
. /workload.sh

function printError() {
    errorString="### $1 ###"
    if [[ $COLORTERMINAL = true ]]; then
        errorString="\033[0;31m$errorString\033[0m"
    fi
    echo -e $errorString
}

function printStep() {
    stepString="[$1]"
    if [[ $COLORTERMINAL = true ]]; then
        stepString="\033[0;36m$stepString\033[0m"
    fi
    echo -e $stepString
}

function printInfo() {
    infoString="> $1"
    if [[ $COLORTERMINAL = true ]]; then
        infoString="\033[0;32m$infoString\033[0m"
    fi
    echo -e $infoString
}

function printOption() {
    echo -e " - '-$1'\t$2"
    for option in $(seq 3 $#); do
        echo -e "   +  $3"        
        shift
    done
}

function determineColors(){
    if [[ $(tput colors) -gt 2 ]]; then
        COLORTERMINAL=true
    fi
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

function usage() {
    echo "Swan's Docker image provides complex solution for building, running and testing swan or running experiment's workloads inside Docker container."
    echo "Usage:"
    printOption "t" "Run selected target. Possible choices are(default: make):" \
        "make" \
        "workload"
    printOption "s" "Selected scenario for target. Possible choices:" \
        "for 'make' target options are specified in swan's Makefile; default: 'integration_test'" \
        "for 'workload' target possible options are: ['caffe', 'memcached', 'mutilate']; default: 'memcached'"
    printOption "p" "Pass parameters to workload binaries. Only for 'workload' target. There is no default parameter."
    printOption "l" "Lock state after executed command has been stopped. Default: false"
}

function parseArguments() {
    printStep "Parsing arguments"
    while getopts "t:s:p:l" opt; do
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
        l)
            LOCKSTATE=true
            ;;
        *)
            usage
            exit
            ;;
        esac
    done
}

function main() {
    determineColors
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
