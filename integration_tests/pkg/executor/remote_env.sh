#!/usr/bin/env bash

# Useful script for integration test in the same directory.

export SWAN_REMOTE_EXECUTOR_TEST_HOST="localhost"
export SWAN_REMOTE_EXECUTOR_USER="root"
export SWAN_REMOTE_EXECUTOR_MEMCACHED_BIN_PATH="/go/src/github.com/intelsdi-x/athena/workloads/data_caching/memcached/memcached-1.4.25/build/memcached"
export SWAN_REMOTE_EXECUTOR_MEMCACHED_USER="memcached"