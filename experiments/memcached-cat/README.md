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

# ![Swan diagram](../../docs/swan-logo-48.png) Swan

## Memcached CAT

This experiment shows best effort jobs behaviour when Memcached load is constant but amount of Level 3 cache and cores assigned to Memcached varies. The number of available caches depends on the platform. See section 17.18 titled "INTEL® RESOURCE DIRECTOR TECHNOLOGY (INTEL® RDT) ALLOCATION FEATURES" of
[Intel® 64 and IA-32 Architectures Software Developer’s Manual](https://software.intel.com/sites/default/files/managed/39/c5/325462-sdm-vol-1-2abcd-3abcd.pdf).

The goal of this experiment is to prove that it is possible to mitigate interference on memory bandwidth and Level 3 cache.

## Caveats

1. Running this experiment requires running a privileged container as [``rdtset``](https://github.com/01org/intel-cmt-cat/tree/master/rdtset) needs to be able to set RMID and COS.
