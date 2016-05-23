#!/usr/bin/env bash

set -e -o pipefail

CACHE_GO=""
GIT_TOKEN_ENV=""


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
    detectDocker
    setCache
    setGitToken
}

function buildingUp() {
    # Skip showing output for clean output on CI
    echo "* Building up ${1} based image"
    docker build -t "swan_${1}_tests" -f Dockerfile_${1} . > /dev/null
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
    setUp
    runBuild "$1"
}

main $1
