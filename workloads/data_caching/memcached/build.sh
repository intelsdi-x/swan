#!/bin/bash
printf "\t\t----------\t\tbuilding memcached\t%s\n" `date +%X`
mkdir -p memcached-1.4.25/build
pushd memcached-1.4.25/build
../configure && make
adduser memcached
popd
printf "\t\t----------\t\tmemcached built\t%s\n" `date +%X`

printf "\t\t----------\t\tbuilding mutilate\t%s\n" `date +%X`
pushd mutilate
scons
popd
printf "\t\t----------\t\tmutilate built\t%s\n" `date +%X`
