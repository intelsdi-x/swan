#!/bin/bash
set -e

CAFFE_DIR=/opt/swan/share/caffe

if [ ! -x ${CAFFE_DIR}/bin/caffe ] ; then
    echo "error: caffe has to be installed $CAFFE_DIR first!"
    exit 1
fi

cd $CAFFE_DIR
export LD_LIBRARY_PATH=/opt/swan/lib:$CAFFE_DIR/lib:$LD_LIBRARY_PATH
./bin/caffe "$@"
