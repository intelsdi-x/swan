# Swan

![Swan diagram](docs/swan-logo.png)

_Scheduler Workloads ANalysis_

[![Build Status](https://travis-ci.com/intelsdi-x/swan.svg?token=EuvqyXrzZzZgasmsv6hn&branch=master)](https://travis-ci.com/intelsdi-x/swan)

## Overview
Swan is a distributed experimentation framework for automated experiments and data collection targeting performance isolation studies. Read Swan [product vision](docs/vision.md) to see what Swan aims for.

Swan use [Snap](https://github.com/intelsdi-x/snap) to collect, process and tag metrics and stores all experiment data in [Cassandra](http://cassandra.apache.org/).
From here, we provide a [Jupyter](http://jupyter.org/) environment to explore and visualize experiment data (more documentation [here](jupyter/README.md)).

Read the [architecture document](docs/architecture.md) to learn more.

![Swan architecture](docs/swan.png)

The first experiment which bundles with Swan is a sensitivity experiment for the distributed
data cache, [memcached](https://memcached.org/). The experiment enables experimenters to generate
a so-called _sensitivity profile_, which describes the violation of _Quality of Service_ under certain conditions, such as CPU cache or network bandwidth interference. An example of the _sensitivity profile_ can be seen below.

![Sensitivity profile](docs/sensitivity-profile.png)

During the experiment *memcached* is colocated with several types of _aggressors_, which are low priority jobs. Memcached response time is critical and needs to stay below a given value which is called _Service Level Objective_ (SLO). SLO is memcached _Quality of Service_ that needs to be maintained. The goal of the experiment is to learn which aggressors interferes the least and which the most with memcached so that some of them can be safely colocated with it without violating memcached _Quality of Service_. _Sensitivity profile_ answers that. Colocation of tasks increases machine utilization which in datacenter [can be low as 12%](https://www.nrdc.org/sites/default/files/data-center-efficiency-assessment-IP.pdf) decreasing _TCO_ of the datacenter. 

Memcached sensitivity experiment is described in detail in [memcached sensitivity profile document](experiments/memcached-sensitivity-profile/README.md).

To read more about general idea behind experiment please refer to the [experiments](docs/experiments.md) document.

## System requirements

Please see [installation guide](docs/install.md#prerequisites) for details.


## Installation

For installation instruction please refer to the [installation guide](docs/install.md).


## Getting started

[Installation guide](docs/install.md) provides step necessary to setup environment and [Sensitivity experiment](experiments/memcached-sensitivity-profile/README.md) document gives instructions how to setup first experiment in the virtual machine.


## Contributing

Best practices for Swan development and submitting code is documented [here](docs/development.md).
