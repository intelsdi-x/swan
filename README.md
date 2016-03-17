# Scheduler Workloads

Repository for automated experiments and data collection targeted enhanced performance isolation and resource oversubscription.

## Instructions

For now, the first supported workload is memcached. Memcached is stressed with the mutilate load generator.
First, you must build memcached and mutilate from source. Go to the [memcached](workloads/data_caching/memcached) workload directory for instructions.

```
$ mkdir -p $GOPATH/src/github.com/intelsdi-x
$ git clone git@github.com:intelsdi-x/swan.git $GOPATH/src/github.com/intelsdi-x/swan
$ cd $GOPATH/src/github.com/intelsdi-x/swan
$ make
```

## Development

When submitting patches, make sure to add test in the pull request and add the test to `./scripts/test.sh` and the new files to `./scripts/lint.sh` for linting.

Before sending or updating pull requests, make sure to run:

```
$ make lint
$ make test
```
