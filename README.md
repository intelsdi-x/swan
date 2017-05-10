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

# Project Swan

![Swan diagram](/images/swan-logo.png)

[![Build Status](https://travis-ci.com/intelsdi-x/swan.svg?token=EuvqyXrzZzZgasmsv6hn&branch=master)](https://travis-ci.com/intelsdi-x/swan)

Table of Contents:
* [Overview](#overview)
* [Quick Start](#quick-start)
* [Terminology](#terminology)
* [System Requirements](#system-requirements)
* [Installation](#installation)
* [Getting Started](#getting-started)
* [Contributing](#contributing)

## Overview
Swan is a distributed experimentation framework for automated experiments targeting performance isolation studies for schedulers. You can read more about the vision behind Swan [here](docs/vision.md).

Swan uses [Snap](https://github.com/intelsdi-x/snap) to collect, process and tag metrics and stores all experiment data in [Cassandra](http://cassandra.apache.org/). From here, we provide a [Jupyter](http://jupyter.org/) environment to explore and visualize experiment data. Read the [architecture document](docs/architecture.md) to learn more.

![Swan architecture](/images/swan.png)

## Quick Start

Swan's Sensitivity Profile Experiment can be quickly run using Vagrant.

Swan requires [VirtualBox](https://www.virtualbox.org/) & [Vagrant](https://www.vagrantup.com/). It is recommended to have [vagrant-vbguest](https://github.com/dotless-de/vagrant-vbguest) plugin installed (`$ vagrant plugin install vagrant-vbguest`).

### Sensitivity Profile Experiment

Run Vagrant image supplied by Swan:

```bash
git clone https://github.com/intelsdi-x/swan
cd swan/vagrant
vagrant plugin install vagrant-vbguest
vagrant box update
vagrant up
vagrant ssh
```

Inside Vagrant, user needs to install Mutilate Load Generator:

```bash
# Please clone the https://github.com/leverich/mutilate repository,
# build it by using `scons` and copy `mutilate` binary to `/bin`.
```

To run experiment, invoke:

```
memcached-sensitivity-profile
```

When experiment is running, please see how to [explore experiment data](TBA) to see results.

While the experiment can be run on developer setup from within a virtual machine or on a laptop, this particular experiment is targeted for  distributed cluster environment. For more details, please see [Memcached Sensitivity Profile Documentation](/experiments/memcached-sensitivity-profile/docs/README.md).

## Terminology

The first experiment which bundles with Swan is a sensitivity experiment for the distributed data cache, [memcached](https://memcached.org/). The experiment allows experimenters to generate a so-called _sensitivity profile_, which describes the violation of _Quality of Service_ under certain conditions, such as CPU cache or network bandwidth interference. An example of the _sensitivity profile_ can be seen below.

![Sensitivity profile](/images/sensitivity-profile.png)

During the experiment *memcached* is colocated with several types of _aggressors_, which are low priority jobs. Memcached response time is critical and needs to stay below a given value which is called _Service Level Objective_ (SLO). SLO is memcached _Quality of Service_ that needs to be maintained. The goal of the experiment is to learn which aggressors interferes the least and which the most with memcached so that some of them can be safely colocated with it without violating memcached _Quality of Service_. _Sensitivity profile_ answers that. Colocation of tasks increases machine utilization which in datacenter [can be low as 12%](https://www.nrdc.org/sites/default/files/data-center-efficiency-assessment-IP.pdf) decreasing _TCO_ of the datacenter.

Memcached sensitivity experiment is described in detail in [memcached sensitivity profile document](experiments/memcached-sensitivity-profile/README.md). To read more about design of an experiment, please refer to the [architecture](docs/architecture.md) document.


## System Requirements

Please see [installation guide](docs/install.md#prerequisites) for details.


## Installation

For installation instruction please refer to the [installation guide](docs/install.md).


## Getting Started

Start to familiarize yourself with Swan by running it locally using the [installation guide](docs/install.md). Once Swan is running, you can begin to experiment. The best place to start is with the [memcached sensitivity experiment](experiments/memcached-sensitivity-profile/README.md).

From there, you can [explore your results with Jupyter](jupyter/README.md).


## Contributing

You can learn how to contribute to the project by checking out the [contributing document](CONTRIBUTING.md). Best practices for Swan development and submitting code is documented [here](docs/development.md).
