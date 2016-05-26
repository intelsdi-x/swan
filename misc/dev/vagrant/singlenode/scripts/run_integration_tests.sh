#!/bin/bash

set -o nounset -o pipefail -o errexit

echo "Preparing environment for integration tests"
/home/vagrant/scripts/start-local-cassandra.sh

# TODO(CD): Export these values to enable remote executor integration tests
# Tracked: SCE-389.
# NB: Those tests are skipped if ANY of these are undefined.
# See integration_tests/pkg/executor/remote_test.go

# export REMOTE_EXECUTOR_TEST_HOST="localhost"
# export REMOTE_EXECUTOR_USER="vagrant"
# export REMOTE_EXECUTOR_SSH_KEY=""
# export REMOTE_EXECUTOR_MEMCACHED_BIN_PATH=""
# export REMOTE_EXECUTOR_MEMCACHED_USER="vagrant"

echo "Running integration test suite"
pushd /home/vagrant/swan
export REMOTE_EXECUTOR_TEST_HOST="localhost"
export REMOTE_EXECUTOR_USER="vagrant"
export REMOTE_EXECUTOR_MEMCACHED_USER="vagrant"
make integration_test
