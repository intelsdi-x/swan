#!/bin/bash
rm -rf memcached-1.4.25/build
mkdir memcached-1.4.25/build
pushd memcached-1.4.25/build
../configure && make
adduser mamcached
popd

pushd mutilate
scons
popd
