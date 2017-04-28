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

# Kubernetes Cluster Launcher

This launcher starts the Kubernetes cluster. It returns a cluster represented as a Task Handle instance.
You can specify two executors:
- One executor specify how to execute master services (and on what host as well) like `apiserver`, `controller-manager` and `scheduler`.
- Second are for minion services like `kubelet` and `proxy`

## Prerequisites

Assumption is that you follow the [installation instructions](docs/install.md) to properly configure this package.

## Note:

It is recommended to use 2 machines. It is important for the accuracy of the Swan experiment to have Kubernetes minion running separated from the master services to avoid interference.
