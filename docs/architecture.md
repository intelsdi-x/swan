# ![Swan diagram](swan-logo-48.png) Swan 

# Architecture overview
Swan is a distributed experimentation framework for automated experiments and data collection targeting performance isolation studies.

Swan use [Snap](https://github.com/intelsdi-x/snap) to collect, process and tag metrics and stores all experiment data in [Cassandra](http://cassandra.apache.org/).
From here, we provide a [Jupyter](http://jupyter.org/) environment to explore and visualize experiment data.

The first experiment which bundles with Swan is a sensitivity experiment for the distributed
data cache, [memcached](https://memcached.org/). The experiment enables experimenters to generate
a so-called sensitivity profile, which describes the violation of Quality of Service under certain
conditions, such as CPU cache or network bandwidth interference. An example of this can be seen below.

![Sensitivity profile](sensitivity-profile.png)

Swan does this by carefully controlling execution of memcached and its co-location with aggressor
processes i.e. noisy neighbors or antagonists. From this point on, Swan coordinates execution of
a distributed load generator called [mutilate](https://github.com/leverich/mutilate).
Snap plugins and tasks are coordinated by swan to collect latency and load metrics and tags them
with experiment identifiers. When the experiment is done, experimenters can fetch the data from
Cassandra or explore the data through Jupyter.


* Experiment overview
* Experiment in details
 * HP job
 * Aggressors
 * Snap
 * Cassandra
 * Jupyter
* Swan's abstractions
 * Launchers
 * Isolations
 * ...
* Sensitivity Profile

 
 
