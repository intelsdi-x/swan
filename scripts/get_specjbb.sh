#!/bin/bash

set -x -e -o pipefail

s3_iso_path=swan-artifacts/SPECjbb2015_1_00.iso
iso_path=workloads/web_serving/SPECjbb2015_1_00.iso
specjbb_path=workloads/web_serving
mnt_path=/mnt/specjbb

yum install -y unzip
wget -P /tmp/ https://github.com/s3tools/s3cmd/archive/master.zip
unzip /tmp/master.zip -d /tmp/
rm /tmp/master.zip
pushd /tmp/s3cmd-master/
python setup.py install
popd

s3cmd sync -c ~/swan_s3_creds/.s3cfg s3://$s3_iso_path $iso_path

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

