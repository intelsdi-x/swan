#!/usr/bin/env bash

CACHE_DIR=".cache"

function createCacheDir() {
    if [[ ! -d $CACHE_DIR ]]; then
        mkdir -p $CACHE_DIR
    fi
}

function buildAndExportImage()  {
    echo "Building ${1} image"
    HASH=$(docker build -t "swan_${1}_tests" -f "integration_tests/docker/Dockerfile_${1}" "integration_tests/docker" | tail -1 | awk '{print $3}')
    HASH_FILE="$CACHE_DIR/swan_$1_tests_hash"

    echo "CURRENT HASH: $HASH"
    echo "HASH FILE: $HASH_FILE"

    if [[ ! -e ${HASH_FILE} || $(cat ${HASH_FILE}) != ${HASH} ]] ; then
        echo "CHANGES DETECTED. NEW HASH IS: $HASH"
        docker save -o "$CACHE_DIR/swan_${1}_tests" "swan_${1}_tests"
        echo ${HASH} > ${HASH_FILE}
    fi
}

function loadExportedImage()    {
    echo "Importing cached image: $1"
    if [[ -e "$CACHE_DIR/swan_${1}_tests" ]]; then
        docker load -i "$CACHE_DIR/swan_${1}_tests"
    fi
}

function main() {
    createCacheDir
    for distro in {"centos","ubuntu"}; do
        loadExportedImage $distro
        buildAndExportImage $distro
    done
}

main
