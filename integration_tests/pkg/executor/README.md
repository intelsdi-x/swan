<!--
 Copyright (c) 2017 Intel Corporation

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

Integration tests prerequisites
===============================

pkg/executor/remote_test.go
---------------------------

This test is intended to pass while connecting to Centos 7 host.

You need to provide following environment variables on your host:
* `REMOTE_EXECUTOR_TEST_HOST` - hostname or IP address of the host that executor should SSH.
* `REMOTE_EXECUTOR_USER` - username to be used to establish SSH connection.
* `REMOTE_EXECUTOR_SSH_KEY` - key (in PEM format) to be used for SSH authentication.
* `REMOTE_EXECUTOR_MEMCACHED_BIN_PATH` - path to memcached binary on the host.
* `REMOTE_EXECUTOR_MEMCACHED_USER` - user to run memcached with (-u option).
