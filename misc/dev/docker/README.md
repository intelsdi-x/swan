# Swan integration with Docker containers

## About

Swan's Docker image provides complex solution for building, running and testing swan or running experiment's workloads inside Docker container.

### Building swan and run integration tests

Swan's image is able to run all Makefile's targets. Due to swan's architecture, container is building and runnning [snap](https://github.com/intelsdi-x/snap)

### Running swan's workloads

Swan's Docker image provides support for following workloads:

- memcached
- mutilate
- caffe

With `-l` parameter container stays running after command execution.

## Building

To build Docker images just run following commands from swan root:

- based on Ubuntu image:

`docker build -t <image_tag> -f ./misc/dev/docker/Dockerfile_ubuntu` .

- based on Centos image:

`docker build -t <image_tag> -f ./misc/dev/docker/Dockerfile_centos` .

where:
- `image_tag` means friendly name for docker image

## Running

To build, test or run swan workload inside Docker container run:

`docker run --privileged -i -t -e GIT_LOGIN=<*your_github_id*> -e GIT_TOKEN=<*your_git_token*> -e GIT_BRANCH=<*target_branch*> -v <*path_to_repo*>:/swan -v /sys/fs/cgroup:/sys/fs/cgroup/:rw --net=host <*image_name*> -t <*target*> -s <*scenario*> -l -p <*params*>`

where:

- `your_github_id` - github account username
- `your_git_token` - per git account token for access to private repositories (if you don't provide it, you will be asked for GitHub credentials during tests)
- `path_to_repo` - absolute path to swan source code (optional)
- `target_branch` - select swan branch for test(s). (default: master)
- `image_name` - image tag which was given during building
- `target` - Run selected target. Possible choices are: 'make', 'command' and 'workload'. (Default: 'make')
- `scenario` - Selected scenario for target. Possible choices are:
    - for 'make' target: options are specified in swan's Makefile (default: 'integration_test')
    - for 'workload' target: \['caffe', 'memcached', 'mutilate', 'l1d', 'l1i', 'l3', 'membw'\] (default: 'memcached')
- `params` - Pass parameters to workload binaries. Only for 'workload' target. (optional)
- `-l` - Don't close container after command execution. (optional)
- `-c` - Custom command which should be run inside container. Only for 'command' target. Default: `bash`.

*Note: If you pass the repository as a volume into container then cloning source code from GitHub will be skipped*

***Warning: Your docker container should be run with following flags:***

- `-v /sys/fs/cgroup:/sys/fs/cgroup/:rw` - this option provides access to cgroups inside container
- `--privileged` - this option provides access to pid namespaces
- `-t` - required by integration tests on `Centos` based image

### Example scenarios:

Run centos based container with running shell. Source code is provided as a volume:

`docker run --privileged -i -t -v <*path_to_repo*>:/swan -v /sys/fs/cgroup:/sys/fs/cgroup/:rw --net=host centos_swan_image`

Run ubuntu based container with caffe workload. Source code will be downloaded from swan's master branch:

`docker run --privileged -i -t -e GIT_TOKEN=<*git_token*> -v /sys/fs/cgroup:/sys/fs/cgroup/:rw --net=host ubuntu_swan_image -t workload -s caffe -p "exp_parameters"`

Run centos based container for unit_test. Source code is provided as a volume:

`docker run --privileged -i -t -v <*path_to_repo*>:/swan -v /sys/fs/cgroup:/sys/fs/cgroup/:rw --net=host centos_swan_image -t make -s unit_test`

Run centos based container with memcached workload. Source code is provided as a volume:

`docker run --privileged -i -t -v <*path_to_repo*>:/swan -v /sys/fs/cgroup:/sys/fs/cgroup/:rw --net=host centos_swan_image -t workload -s memcached`

Run centos based container with custom command. Source code is provided as a volume:

`docker run --privileged -i -t -v <*path_to_repo*>:/swan -v /sys/fs/cgroup:/sys/fs/cgroup/:rw --net=host centos_swan_image -t command -c /bin/bash`
