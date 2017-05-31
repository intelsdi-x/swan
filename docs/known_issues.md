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

# Known Issues

## Using Kernel 4.11 and later

RDMA cgroup controller was introduced in version 4.11 of Linux. Kubernetes 1.6 (as of 31st of May 2016) does not support the controller and will prevent `kubelet` from launching. It is necessary to patch Kubernetes and build hyperkube. Following snippet can be used:

```sh
go get k8s.io/kubernetes
git checkout 1.6.4 
git cherry-pick 4e002eacb11777ddfc06c2f4e361e7c453ab4d9d
make all WHAT="cmd/hyperkube" KUBE_BUILD_PLATFORMS="linux/amd64"
```

Compiled binary can be found at `_output/bin/hyperkube`. You should distribute it to all the Kubernetes nodes used for experimentation.

The issue will be fixed in Kubernetes 1.7.
