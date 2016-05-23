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

**Note:** Your GIT_TOKEN variable(which contains private token string for access to GitHub repositories) is automatically pass into containers. 

# Integration tests in Docker Container

## Building

To build Docker images just run:

- based on Ubuntu image:
```sh
docker build -t <image_tag> ./integration_tests/docker/ubuntu/
```
- based on Centos image:
```sh
docker build -t <image_tag> ./integration_tests/docker/centos/
```

where:
- `image_tag` means friendly name for docker image

## Running

To run integration tests inside Docker container run:
```sh
docker run --privileged -i -t -e GIT_TOKEN=<*your_git_token*> -e GIT_BRANCH=<*target_branch*> -v <*path_to_repo*>:/swan -v /sys/fs/cgroup:/sys/fs/cgroup/:rw --net=host <*image_name*> <*target*>
```
where:
- `your_git_token` - per git account token for access to private repositories (if you don't provide it, you will be asked for GitHub credentials during tests)
- `path_to_repo` - absolute path to swan source code (optional)
- `target_branch` - select swan branch for test(s). (default: master)
- `image_name` - image tag which was given during building
- `target` - list of targets, which should be run by docker (default: integration_test)

*Note: If you pass repository as a volume into container then cloning source code from GitHub will be skipped*

***Warning: Your docker container should be run with following flags:***
- `-v /sys/fs/cgroup:/sys/fs/cgroup/:rw` - this option provides access to cgroups inside container
- `--privileged` - this option provides access to pid namespaces 
- `-t` - required by integration tests on `Centos` based image
