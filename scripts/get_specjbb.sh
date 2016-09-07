#!/bin/bash

set -x -e -o pipefail

iso_path=workloads/web_serving/SPECjbb2015_1_00.iso
specjbb_path=workloads/web_serving
mnt_path=/mnt/specjbb

# Download SPECjbb here, when we will have decision where to put it.

if [ -e $iso_path ]
then
    mkdir -p $mnt_path
    mount -o loop $iso_path $mnt_path
    mkdir -p $specjbb_path
    cp -R $mnt_path $specjbb_path
    umount $mnt_path
    chmod -R +w $specjbb_path/specjbb/config
else
    echo Could not find SPECjbb ISO file \($iso_path\)
fi

