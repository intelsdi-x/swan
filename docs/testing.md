# Integration tests

Integration tests are run separately from unit_tests. This is cause they require custom
configuration and setup like special packages and building the workload binary.

# Setup steps

## Setup for Memcached Workload tests

Before test, make sure:
- The Memcached is built.
  - Go to the [memcached](workloads/data_caching/memcached) workload directory for instructions.
  - Optionally, export MEMCACHED_BIN variable if you want to use memcached in custom path.
- User `memcached` is present.
- `libevent` package is present.
- `nc` program is present.
  - Centos `yum install nc`
  - Ubuntu `apt-get netcat`

## Setup for Isolation tests

Before test, make sure:
- Install `cgroup tools`
  - Centos `yum install libcgroup libcgroup-tools`
  - Ubuntu `apt-get install libcgroup-dev cgroup-bin`

# Using with go test

After setup you can run them in following manner:

`go test -tags=integration`

To create integration test file we use build tags, so you need to place

```
// +build integration

package integration
```

NOTE: Make sure you place newline between package name and build flag

# Using with makefile

After setup you can run unit tests only in following manner:

`make unit_test`

To run all tests including integration tests:

`make test`

To run integration tests inside Docker containers:

`make integration_test_on_docker`

**Note:** Optionally, you can set GIT_TOKEN variable to get private GitHub repositories used in this test (variable will be passed into containers automatically).

# Using docker containers

To run integration tests in docker containers please follow instruction from [Swan integration with docker containers](../misc/dev/docker/README.md)
