#!/usr/bin/env bash

set -e -o pipefail

CACHE_GO=""
GIT_TOKEN_ENV=""
DATE=`date`


function detectDocker() {
    if [[ $(which docker) == "" ]]; then
        echo "Docker hasn't been detected. Skipping"
        exit 0
    fi
}

function setCache() {
    if [[ -d "$(pwd)/../../.cache" ]]; then
        CACHE_GO="-v $(pwd)/../../.cache/gopath/:/opt/gopath"
    fi
}

function setGitToken() {
    if [[ ${GIT_TOKEN} != "" ]]; then
        GIT_TOKEN_ENV="-e GIT_TOKEN=${GIT_TOKEN}"
    fi
}

function setUp() {
    printf "\t\tdetectDocker\t%s\n" `date +%X`
    detectDocker
    printf "\t\tsetCache\t%s\n" `date +%X`
    setCache
    printf "\t\tsetGitToken\t%s\n" `date +%X`
    setGitToken
    printf "\t\tsetUp done\t%s\n" `date +%X`
}

function buildingUp() {
    # Skip showing output for clean output on CI
    echo "* Building up ${1} based image"
    docker build -t "swan_${1}_tests" -f Dockerfile_${1} . 
}

function running() {
    echo "* Running ${1} based image"
    docker run --privileged $GIT_TOKEN_ENV $CACHE_GO -t -v $(pwd)/../../:/swan -v /sys/fs/cgroup:/sys/fs/cgroup:rw --net=host "swan_${1}_tests"

}

function selectCases(){
    case $1 in
        ubuntu )
            echo "ubuntu";;
        centos )
            echo "centos";;
        *)
            echo "centos ubuntu"
    esac
}

function runBuild() {
    for buildCase in $(selectCases $1) ; do
        echo "### ${buildCase} TEST ###"
        buildingUp $buildCase
        running $buildCase
    done
}

function main() {
    printf "\t\tsetUp start\t%s\n" `date +%X`
    setUp
    printf "\t\trunBuild\t%s\n" `date +%X`
    runBuild "$1"
    printf "\t\tmain done\t%s\n" `date +%X`
}

main $1
