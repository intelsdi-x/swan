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

## Development

Before sending or updating pull requests, make sure to run:

test & build & run
```
$ make deps
$ make              # lint test build
$ make run
```

### or just in simple go way
```
go test ./pkg/...
golint ./pkg/...
go build ./cmds/memcache
```

### Depedency managment

Handled by [godeps](https://github.com/tools/godep).
