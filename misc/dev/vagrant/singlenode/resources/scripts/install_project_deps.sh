#!/bin/bash 

set -e

. $HOME_DIR/.bash_profile

function executeAsUser() {
        sudo -E -u $VAGRANT_USER -s PATH=$PATH GOPATH=$GOPATH CCACHECONFDIR=$CCACHECONFDIR "$@"
}

echo "Installing project dependencies..."
pushd $HOME_DIR/go/src/github.com/intelsdi-x/swan/
executeAsUser make repository_reset
executeAsUser make deps_all
if [[ "$BUILD_AMI" == "true" ]]; then
        executeAsUser make build_image
fi
executeAsUser make build_workloads
popd
