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
executeAsUser make build_workloads

# -b specifies bucket name.
# By default bucket name value is read from SWAN_BUCKET_NAME env variable.
# When we add this variable to jenkins/vagrant, we will be able to remove it from command below.
./scripts/get_specjbb.sh -s . -c $HOME_DIR/swan_s3_creds/.s3cfg -b swan-artifacts
popd
