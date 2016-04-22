# Scheduler Workloads

[![Build Status](https://travis-ci.com/intelsdi-x/swan.svg?token=EuvqyXrzZzZgasmsv6hn&branch=master)](https://travis-ci.com/intelsdi-x/swan)

Repository for automated experiments and data collection targeted enhanced performance isolation and resource oversubscription.

## Instructions

For now, the first supported workload is memcached. Memcached is stressed with the mutilate load generator.
First, you must build memcached and mutilate from source. Go to the [memcached](workloads/data_caching/memcached) workload directory for instructions.

```
$ go get github.com/intelsdi-x/swan
$ make deps
$ make
```

## Development using Makefile

Before sending or updating pull requests, make sure to run:

test & build & run
```
$ make deps
$ make              # lint unit_test build
$ make run
```

### Detailed options for tests
```
$ make test TEST_OPT="-v -run <specific test>"
```

### Development using go binaries
```
go test ./pkg/...
golint ./pkg/...
go build ./cmds/memcache
```

### Depedency managment

Handled by [godeps](https://github.com/tools/godep).

### Integration tests

For Swan Workload integration tests see [README](src/pkg/workloads/integration/README.md) file for instructions.

### Mock generation

Mock generation is done by Mockery tool.
Sometimes Mockery is not able to resolve all imports in file correctly.
Developer needs to use it manually, that's why we are vendoring our mocks.

To generate mocks go to desired package and ```mockery -name ".*" -case underscore```
