#!/bin/bash

set -e -x

SNAP_VERSION="v1.0.0"

. $HOME_DIR/.bash_profile
ATHENA_DIR=$GOPATH/src/github.com/intelsdi-x/athena

echo "Installing Snap..."
wget https://s3-us-west-2.amazonaws.com/snap.ci.snap-telemetry.io/snap/1.0.0/linux/x86_64/snaptel -O $HOME_DIR/go/bin/snaptel
chmod +x $HOME_DIR/go/bin/snaptel

wget https://s3-us-west-2.amazonaws.com/snap.ci.snap-telemetry.io/snap/1.0.0/linux/x86_64/snapteld -O $HOME_DIR/go/bin/snapteld
chmod +x $HOME_DIR/go/bin/snapteld

wget https://s3-us-west-2.amazonaws.com/snap.ci.snap-telemetry.io/snap/1.0.0/linux/x86_64/snap-plugin-collector-mock1 -O $HOME_DIR/go/bin/snap-plugin-collector-mock1
chmod +x $HOME_DIR/go/bin/snap-plugin-collector-mock1

wget https://s3-us-west-2.amazonaws.com/snap.ci.snap-telemetry.io/snap/1.0.0/linux/x86_64/snap-plugin-publisher-mock-file -O $HOME_DIR/go/bin/snap-plugin-publisher-mock-file
chmod +x $HOME_DIR/go/bin/snap-plugin-publisher-mock-file

echo "Installing Athena & its K8s..."
if [ ! -d $ATHENA_DIR ]; then
    echo "Fetching Athena sources"
    mkdir -p $ATHENA_DIR
    git clone git@github.com:intelsdi-x/athena $ATHENA_DIR
else
    echo "Updating Athena sources"
    pushd $ATHENA_DIR
    git pull
    popd
fi
echo "Fetching kubernetes binaries for Athena"
cd $ATHENA_DIR && ./misc/kubernetes/install_binaries.sh
