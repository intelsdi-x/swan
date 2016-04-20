# Workload integration tests

In this directory there all integration tests for each workload. These tests requires some
setup before execution like special packages and building the workload binary.

# Setup for Memcached tests

Before test, make sure:
- The Memcached is built.
  - Go to the [memcached](workloads/data_caching/memcached) workload directory for instructions.
  - Optionally, export MEMCACHED_BIN variable if you want to use memcached in custom path.
- User `memcached` is present.
- `nc` program is present. (`apt-get netstat` or `yum install nc`)

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