# Swan integration with Docker containers

## Building

To build Docker images just run:

- based on Ubuntu image:

`docker build -t <image_tag> ./integration_tests/docker/ubuntu/`

- based on Centos image:

`docker build -t <image_tag> ./integration_tests/docker/centos/`


where:
- `image_tag` means friendly name for docker image

## Running

To run integration tests inside Docker container run:

`docker run --privileged -i -t -e GIT_TOKEN=<*your_git_token*> -e GIT_BRANCH=<*target_branch*> -v <*path_to_repo*>:/swan -v /sys/fs/cgroup:/sys/fs/cgroup/:rw --net=host <*image_name*> -t <*target*> -s <*scenario*> -p <*params*>`

where:

- `your_git_token` - per git account token for access to private repositories (if you don't provide it, you will be asked for GitHub credentials during tests)
- `path_to_repo` - absolute path to swan source code (optional)
- `target_branch` - select swan branch for test(s). (default: master)
- `image_name` - image tag which was given during building
- `target` - Run selected target. Possible choices are: 'make' and 'workload'. (Default: 'make')
- `scenario` - Selected scenario for target. Possible choices are:
    - for 'make' target: options are specified in swan's Makefile (default: 'integration_test')
    - for 'workload' target: \['caffe', 'memcached', 'mutilate'\] (default: 'memcached')
- `params` - Pass parameters to workload binaries. Only for 'workload' target. (optional)

*Note: If you pass repository as a volume into container then cloning source code from GitHub will be skipped*

***Warning: Your docker container should be run with following flags:***

- `-v /sys/fs/cgroup:/sys/fs/cgroup/:rw` - this option provides access to cgroups inside container
- `--privileged` - this option provides access to pid namespaces
- `-t` - required by integration tests on `Centos` based image
