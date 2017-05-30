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

## Example results of Memcached CAT experiment.

### Application performance metrics.

#### SLI violation table.

![SLI table](latency-table.png)

#### Raw latency normalized to SLO.

![Latency normalized](hp_latency_normalized.png)

#### Raw latency absolute.

![Latency absolute](hp_latency_absolute_us.png)

### Platform Intel RDT (CMT/MBM) metrics.

#### memcached (HighPriority) LastLevelCache Occupancy (MB)

![HP LLC Occupancy](hp_llc_occupancy_mb.png)

#### Aggressors LastLevelCache Occupancy (MB)

![BE LLC Occupancy](be_llc_occupancy_mb.png)

#### memcached (High Priority) memory bandwidth (GB)

![HP memory bandwidth](hp_membw_gb.png)

#### Aggressors memory bandwidth (GB)

![BE memory_bandwidth](be_membw_gb.png)
