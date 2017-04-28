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

# Prerequisites

While the experiment can be run on developer setup from within a virtual machine or on a laptop, it is targeted for the data center environment with high bandwidth links and rich multi socket servers. Therefore further steps will assume that experiment will be run in the multi node environment.

## Machine configuration

When Swan is used in a cluster environment, we recommend the following machine topology:

| Type                  | Description                                                                                                                               | Machine                                                                                |
|-----------------------|-------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|
| System Under Test (SUT)        | Machine where Swan is run and thus where workloads are deployed.                             | 1 x 10Gb link, hyper threaded with 16 or more hyper threads, preferably with 2 sockets |
| Load generator agents | Machines to generate stress on the target machine. Our application of choice for this task is [Mutilate](https://github.com/leverich/mutilate).                                                                                       | 10Gb link for each agent, 20 or more hyper threads in total                                       |
| Services node         | Machine where Cassandra, Jupyter and Load Generator Master will run. The 'cleaniness' of this machine is less important than target and load generator machines. | 1 x 1-10Gb link, higher memory capacity to accommodate for Cassandra heap usage.       |

When Memcached is run on large multi-core machine, then it will require multiple load generator agents be fully saturated.

## Hardware

Although Swan can be run on any machine, the more recent hardware might yield more detailed metrics about workload interference. For example Haswell CPU (Xeon v3) enables monitoring of cache occupancy (via CMT). Newer Broadwell CPU (Xeon v4) adds Memory Bandwidth Monitoring and enables user to separate cache for different workloads (via CAT). For more information, please see [Intel RDT page](http://www.intel.com/content/www/us/en/architecture-and-technology/resource-director-technology.html).
 
## Software

Swan was tested on CentOS 7, and installation instructions covers only this distribution, although other distribution should be able to run Swan just fine.

In the next section there is a list of default options of CentOS 7 kernel that should be changed to accommodate larger loads and make latency variance lower.
Experiment when run will log `Warning` messages if those setting are not set properly, but will run nonetheless.

### File descriptors
This should be set on SUT and load generator machines.

As the both Mutilate and Memcached will create many connections, it is important that the number of available file descriptors is high enough. It should be in the high tens of thousands.
To check the current limit, run:

```bash
$ ulimit -n
256
```

and set a new value with:

```bash
$ ulimit -n 65536
```

### DDoS protection
This should be set on SUT machine.

Sometimes, the Linux kernel applies anti-denial of service measures, like introducing [TCP SYN cookies](https://en.wikipedia.org/wiki/SYN_cookies). This will break the mutilate load generators and should be turned off on the SUT machine:

```bash
$ sudo sysctl net.ipv4.tcp_syncookies=0
```

### Power control (Bare Metal only)
This should be set on SUT machine. Does not affect virtual machines.

To avoid power saving policies to kick in while carrying out the experiments, set the power governor policy to 'performance':

```bash
$ sudo cpupower frequency-set -g performance
```

## Next
Please move to [Installation](installation.md) page.
