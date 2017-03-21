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
