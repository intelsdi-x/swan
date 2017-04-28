<!--
 Copyright (c) 2017 Intel Corporation

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# ![Swan logo](/images/swan-logo-48.png) Swan 

# Integration with Docker containers

## About

Swan's Docker image provides complex solution for building and running swan workloads inside Docker container.

### Running swan's workloads

Swan's Docker image provides support for following workloads:

- memcached
- mutilate
- caffe
- low_level_aggressors
- stream

## Building

To build Docker images just run following commands from swan root:

- based on Centos image:

`docker build -t <image_tag> -f ./misc/dev/docker/Dockerfile` .

where:
- `image_tag` means friendly name for docker image

## Running

To build, test or run swan workload inside Docker container run:

`docker run --privileged -i -t -v /sys/fs/cgroup:/sys/fs/cgroup/:rw --net=host <*image_name*>`

where:

- `image_name` - image tag which was given during building

***Warning: Your docker container should be run with following flags:***

- `-v /sys/fs/cgroup:/sys/fs/cgroup/:rw` - this option provides access to cgroups inside container
- `--privileged` - this option provides access to pid namespaces
- `-t` - required by integration tests on `Centos` based image
