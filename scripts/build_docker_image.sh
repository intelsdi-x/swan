#!/bin/bash


SWAN_DIR="$(dirname ${BASH_SOURCE[0]})/.."
FILENAME=$(cat ${SWAN_DIR}/latest_build)

cp ${SWAN_DIR}/${FILENAME} ${SWAN_DIR}/artifacts.tar.gz
pushd ${SWAN_DIR}
docker build -t centos_swan_image .
popd
