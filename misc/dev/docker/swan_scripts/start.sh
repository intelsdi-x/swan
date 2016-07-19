#!/bin/bash
set -o pipefail

REPOSITORY_URL=github.com/intelsdi-x/swan

TARGET=""
SCENARIO=""
BINPARAMETERS=""
LOCKSTATE=false
COLORTERMINAL=false
GIT_REPO_LOCATION=""
DECORATOR=""

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
    if [[ $GIT_TOKEN != "" && $GIT_LOGIN != "" ]]; then
        echo -e "machine github.com\nlogin $GIT_LOGIN\npassword $GIT_TOKEN" > ~/.netrc
        printInfo "Token for GitHub has been set for user $GIT_LOGIN"
    fi
}

function setRemoteRepo() {
    printStep "Setting up remote repository"
    
    if [[ $GIT_BRANCH == "" ]]; then
        GIT_BRANCH="master"
    fi
       
    verifyStatus
    printInfo "Remote repository configuration has been set"
}

function setLocalRepo() {
    printStep "Setting up local source repository"

    pushd /swan
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    popd
    git remote remove $GIT_REPO_LOCATION &> /dev/null
    git remote add $GIT_REPO_LOCATION /swan/.git
    verifyStatus
    printInfo "Local repository configuration has been set"
}

function prepareEnvironment() {
    printStep "Prepare environment"
    make deps
    verifyStatus
    printInfo "All dependencies have been downloaded"
}

function getCode() {
    cd swan
    if [[ -d "/swan" ]]; then
        GIT_REPO_LOCATION="local_repo"
        setLocalRepo
    else
        GIT_REPO_LOCATION="origin"
        setRemoteRepo
    fi
    git pull $GIT_REPO_LOCATION $GIT_BRANCH
    git checkout -b $GIT_REPO_LOCATION/$GIT_BRANCH
    cd ..
}

function usage() {
    echo "Swan's Docker image provides complex solution for building, running and testing swan or running experiment's workloads inside Docker container."
    echo "Usage:"
    printOption "t" "Run selected target. Possible choices are(default: make):" \
        "make" \
        "workload"
    printOption "s" "Selected scenario for target. Possible choices:" \
        "for 'make' target options are specified in swan's Makefile; default: 'integration_test'" \
        "for 'workload' target possible options are: ['caffe', 'memcached', 'mutilate', 'l1d', 'l1i', 'l3', 'membw']; default: 'memcached'"
    printOption "p" "Pass parameters to workload binaries. Only for 'workload' target. There is no default parameter."
    printOption "l" "Lock state after executed command has been stopped. Default: false"
    printOption "d" "Decorate workload with custom command. Only for 'workload' target. Empty option doesn't set decorator. Default: \"\""
}

function parseArguments() {
    printStep "Parsing arguments"
    while getopts "t:s:p:d:l" opt; do
    case $opt in
        t)
            TARGET=$OPTARG
            ;;
        s)
            SCENARIO=$OPTARG
            ;;
        d)
            DECORATOR="$OPTARG"
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
