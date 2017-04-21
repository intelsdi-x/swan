# Copyright (c) 2017 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/usr/bin/env bash

# Useful script for integration test in the same directory.

export SWAN_REMOTE_EXECUTOR_TEST_HOST="localhost"
export SWAN_REMOTE_EXECUTOR_USER="root"
export SWAN_REMOTE_EXECUTOR_MEMCACHED_BIN_PATH="/go/src/github.com/intelsdi-x/swan/workloads/data_caching/memcached/memcached-1.4.25/build/memcached"
export SWAN_REMOTE_EXECUTOR_MEMCACHED_USER="memcached"