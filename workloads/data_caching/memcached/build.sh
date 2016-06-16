#!/bin/bash
set -e
mkdir -p memcached-1.4.25/build
pushd memcached-1.4.25/build
../configure && make
id -u memcached &>/dev/null || sudo adduser memcached
popd

pushd mutilate
scons
popd
