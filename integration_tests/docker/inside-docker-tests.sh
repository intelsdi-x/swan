#!/usr/bin/env bash

if [[ $(which docker) == "" ]]; then
    echo "Docker hasn't been detected. Skipping"
    exit 0
fi

echo "Building up docker images"
# Skip showing output for clean output on CI
echo "* Building up Centos based image"
docker build -t swan_centos_tests ./centos/ > /dev/null
echo "* Building up Ubuntu based image"
docker build -t swan_ubuntu_tests ./ubuntu/ > /dev/null

echo "Running up tests"
echo "* Running Centos based image"
docker run --privileged -t -v $(pwd)/../../:/swan -v /sys/fs/cgroup:/sys/fs/cgroup:rw --net=host swan_centos_tests
if [[ $? -gt 0 ]]; then
    exit 1
fi
echo "* Running Ubuntu based image"
docker run --privileged -t -v $(pwd)/../../:/swan -v /sys/fs/cgroup:/sys/fs/cgroup:rw --net=host swan_ubuntu_tests

