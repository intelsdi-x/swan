#!/bin/bash

set -x -e -o pipefail

s3_iso_path=swan-artifacts/SPECjbb2015_1_00.iso
specjbb_path=workloads/web_serving
iso_path=$specjbb_path/SPECjbb2015_1_00.iso
mnt_path=/mnt/specjbb

pip install s3cmd==1.6.1
pushd $HOME_DIR/go/src/github.com/intelsdi-x/swan/
s3cmd sync -c $HOME_DIR/swan_s3_creds/.s3cfg s3://$s3_iso_path $iso_path

if [ -e $iso_path ]
then
    mkdir -p $mnt_path
    mount -o loop $iso_path $mnt_path
    cp -R $mnt_path $specjbb_path
    umount $mnt_path
    chmod -R +w $specjbb_path/specjbb/config
else
    echo Could not find SPECjbb ISO file \($iso_path\)
fi

popd

