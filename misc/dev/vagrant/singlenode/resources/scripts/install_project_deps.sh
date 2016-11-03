#!/bin/bash 

set -e

. $HOME_DIR/.bash_profile

function executeAsVagrantUser() {
        sudo -E -u $VAGRANT_USER -s PATH=$PATH GOPATH=$GOPATH CCACHECONFDIR=$CCACHECONFDIR "$@"
}

BUILD_OPENBLAS=""

echo "Installing project dependencies..."
pushd $HOME_DIR/go/src/github.com/intelsdi-x/swan/
executeAsVagrantUser make repository_reset
executeAsVagrantUser make deps_all
if [[ "$BUILD_DOCKER_IMAGE" == "true" ]]; then
        executeAsVagrantUser make build_image
        BUILD_OPENBLAS="true"
fi
executeAsVagrantUser make BUILD_OPENBLAS=${BUILD_OPENBLAS} build_workloads

# -b specifies bucket name.
# By default bucket name value is read from SWAN_BUCKET_NAME env variable.
# When we add this variable to jenkins/vagrant, we will be able to remove it from command below.
./scripts/get_specjbb.sh -s . -c $HOME_DIR/swan_s3_creds/.s3cfg -b swan-artifacts
popd
