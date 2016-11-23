#!/bin/bash
set -e
mkdir -p memcached-1.4.25/build
pushd memcached-1.4.25/build
../configure && make --quiet
id -u memcached &>/dev/null || sudo adduser memcached
popd

pushd mutilate
rm -rf .sconf_temp
git clean -fdx
scons
popd
