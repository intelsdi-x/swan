#!/usr/bin/env bash

CACHE_DIR=".cache"

function createCacheDir() {
    if [[ ! -d $CACHE_DIR ]]; then
        mkdir $CACHE_DIR
    fi
}

function buildAndExportImage()  {
    docker build -t "swan_${1}_tests" -f "integration_tests/docker/Dockerfile_${1}" "integration_tests/docker"
    docker save -o "$CACHE_DIR/swan_${1}_tests" "swan_${1}_tests"
}

function loadExportedImage()    {
    if [[ -a "$CACHE_DIR/swan_${1}_tests" ]]; then
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
