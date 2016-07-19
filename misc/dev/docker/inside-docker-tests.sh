#!/usr/bin/env bash

set -e -o pipefail

if [[ $(which docker) == "" ]]; then
    echo "Docker hasn't been detected. Skipping"
    exit 0
fi

GIT_TOKEN_ENV=""
if [[ ${GIT_TOKEN} != "" ]]; then
    GIT_TOKEN_ENV="-e GIT_TOKEN=${GIT_TOKEN}"
fi

cd ../../../

echo "Building up docker images"
# Skip showing output for clean output on CI
echo "* Building up Centos based image"
docker build -t misc/dev/docker/swan_centos_tests -f Dockerfile_centos . > /dev/null
echo "* Building up Ubuntu based image"
docker build -t misc/dev/docker/swan_ubuntu_tests -f Dockerfile_ubuntu . > /dev/null

echo "Running up tests"
echo "* Running Centos based image"
docker run --privileged $GIT_TOKEN_ENV -t -v $(pwd):/swan -v /sys/fs/cgroup:/sys/fs/cgroup:rw --net=host swan_centos_tests
echo "* Running Ubuntu based image"
docker run --privileged $GIT_TOKEN_ENV -t -v $(pwd):/swan -v /sys/fs/cgroup:/sys/fs/cgroup:rw --net=host swan_ubuntu_tests

