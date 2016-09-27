#!/bin/bash

set -e

pip install s3cmd==1.6.1

usage() {
    echo "$(basename ${0}) [-s swan_path] [-c credentials_file] [-b bucket]"
    echo " - a script to download and extract SPECjbb.iso from S3."
    echo ""
    echo "Options:"
    echo " -s swan_path: Path to swan repository (default ${HOME}/go/src/github.com/intelsdi-x/swan)"
    echo " -c credentials_file: Path to file with S3 credentials (default ${HOME}/swan_s3_creds/.s3cfg)"
    echo " -b bucket: S3 Bucket name (default ${SWAN_BUCKET_NAME})"
    echo " -h: to see this help"
}

SWAN_PATH=$HOME/go/src/github.com/intelsdi-x/swan
S3_CREDS_FILE=$HOME/swan_s3_creds/.s3cfg

download() {
    echo "Downloading SPECjbb iso file from S3."
    S3_ISO_PATH=$SWAN_BUCKET_NAME/SPECjbb2015_1_00.iso
    s3cmd sync -c $S3_CREDS_FILE s3://$S3_ISO_PATH $SPECJBB_ISO_PATH
}

extract() {
    echo "Extracting SPECjbb iso file to $SPECJBB_PATH"
    MNT_PATH=/mnt/specjbb
    if [ -e $SPECJBB_ISO_PATH ]; then
        mkdir -p $MNT_PATH
        mount -o loop $SPECJBB_ISO_PATH $MNT_PATH
        cp -R $MNT_PATH $SPECJBB_PATH
        umount $MNT_PATH
        chmod -R +w $SPECJBB_PATH/specjbb/config
        echo "SPECjbb files extracted to $SPECJBB_PATH/specjbb"
    else
        echo "Could not find SPECjbb ISO file ($SPECJBB_ISO_PATH)."
    fi
}

while getopts "hs:c:b:" OPT; do
    case "$OPT" in
        h)
            usage
            exit 0
            ;;
        s)
            SWAN_PATH="${OPTARG}"
            ;;
        c)
            S3_CREDS_FILE="${OPTARG}"
            ;;
        b)
            SWAN_BUCKET_NAME="${OPTARG}"
            ;;
    esac
done

SPECJBB_PATH=$SWAN_PATH/workloads/web_serving
SPECJBB_ISO_PATH=$SPECJBB_PATH/SPECjbb2015_1_00.iso

if [ -z "$SWAN_BUCKET_NAME" ]; then
    echo "Please provide your S3 bucket name"
    exit 1
fi

download
extract
