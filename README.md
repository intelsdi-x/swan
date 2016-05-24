# Scheduler Workloads

[![Build Status](https://travis-ci.com/intelsdi-x/swan.svg?token=EuvqyXrzZzZgasmsv6hn&branch=master)](https://travis-ci.com/intelsdi-x/swan)

Repository for automated experiments and data collection targeted enhanced performance isolation and resource oversubscription.

## Instructions

For now, the first supported workload is memcached. Memcached is stressed with the mutilate load generator.
First, you must build memcached and mutilate from source. Go to the [memcached](workloads/data_caching/memcached) workload directory for instructions.

**Local development**

```
$ go get github.com/intelsdi-x/swan
$ make deps
$ make
```

**Vagrant (Virtualbox) development environment**

Follow the [Vagrant instructions](misc/dev/vagrant/singlenode/README.md) to
create a Linux virtual machine pre-configured for developing Swan.

## Contributing

Best practices for Swan development and submitting code is documented [here](docs/development.md).
