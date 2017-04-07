# Theory

## Workload Interference

As the experiments measures sub-millisecond response times, there are a myriad of sources of interference which silently can cause misleading measurements.
To get insight into some of these, please refer to [Kozyrakis, Jacob Leverich Christos. "Reconciling High Server Utilization and Sub-millisecond Quality-of-Service"](http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.713.5120&rep=rep1&type=pdf).

Much of the configuration guidelines here are targeted eliminating as many of these (unintentional) sources of interference as possible.

Swan has built-in performance isolation patterns to focus aggressors on the sources of interference they are intended to stress.
However, Swan needs some input from the user about the environment to adjust these. The sections below will go over the recommended settings.

### CPU Cores Isolation

To give idea why and how to isolate tasks, first please take a look at simplified graph showing CPU topology. In this example there is a *n* core physical CPU with *HyperThreading* enabled. Each core has two execution threads. In the Linux system each execution thread is reported as logical CPU and without knowing the CPU topology user cannot easily guess which logical CPUs are execution threads in the same core.

![Cache topology](../../../docs/cpu_topo.png)

`lstopo` and `lscpu` can help to understand the CPU topology in a given system. `lstopo` will show CPU topology graphically while `lscpu -e` will show CPU topology in text mode giving much more details:
```bash
$ lscpu -e
CPU NODE SOCKET CORE L1d:L1i:L2:L3 ONLINE MAXMHZ    MINMHZ
0   0    0      0    0:0:0:0       yes    4000,0000 800,0000
1   0    0      1    1:1:1:0       yes    4000,0000 800,0000
2   1    1      2    2:2:2:0       yes    4000,0000 800,0000
3   1    1      3    3:3:3:0       yes    4000,0000 800,0000
4   0    0      0    0:0:0:0       yes    4000,0000 800,0000
5   0    0      1    1:1:1:0       yes    4000,0000 800,0000
6   1    1      0    2:2:2:0       yes    4000,0000 800,0000
7   1    1      3    3:3:3:0       yes    4000,0000 800,0000

```
In the exemplary output from the `lscpu -e` there is a dual socket platform with 2 physical CPUs each of which has 2 cores and *HyperThreading* enabled. Output shows that:

* Logical CPUs 0 and 4 are execution threads on core 0 on the socket 0
* Logical CPUs 1 and 5 are execution threads on core 1 on the socket 0
* Logical CPUs 2 and 6 are execution threads on core 0 on the socket 1
* Logical CPUs 3 and 7 are execution threads on core 0 on the socket 1
* Logical CPUs 0 and 4 share L1 instruction, L1 data and L2 cache
* Logical CPUs 1 and 5 share L1 instruction, L1 data and L2 cache
* Logical CPUs 0, 1, 4 and 5 share L3 cache
* Logical CPUs 1, 2, 4, 5 reside on a different socket and NUMA node than Logical CPUs 2, 3, 6, 7. They have separate caches and memory access to RAM.

The Linux scheduler detects 8 CPUs and in spite of the fact that it's taking into account the CPU topology it may schedule jobs in a way that they will work in a very inefficient way. It may even migrate jobs between NUMA nodes increasing job's memory access time. Please read [The Linux Scheduler: a Decade of Wasted Cores](https://www.ece.ubc.ca/~sasha/papers/eurosys16-final29.pdf) to learn more.


To give insight into the placement of aggressor workloads, and motivate the thread count selection in memcached, let us start with an example topology:

![Empty topology](../../../docs/topology-1.png)

Using half the number of physical cores on one socket leaves us with 1 memcached thread:

![Memcached topology](../../../docs/topology-2.png)

We do this, partly so we can introduce isolated aggressors on the L1 caches:
![Memcached + L1 topology](../../../docs/topology-3.png)

_and_ introduce L3 aggressors with the same setup of memcached, in order to compare latency measurements between both aggressor types:

![Memcached + L3 topology](../../../docs/topology-4.png)

Swan will by default try to aim for the core configuration above via instrumenting Linux scheduler to run each workload on specific cores. 

Using exclusive CPU sets can be challenging if other systems on the host are using CPU sets. Exclusive CPU sets cannot share cores with any other cgroup and setting the desired cores will cause an error from the kernel.
An example of such conflicting and potential overlapping CPU sets could be systems with [docker](https://www.docker.com/) installed. Docker creates a cpuset cgroup which contain all logical cores and thus will conflict with Swan, if Swan attempts to create exclusive CPU sets.

## Synthetic Aggressors

Synthetic Aggressors are specialized programs for stressing different platform subsystems.

| Source of interference | Aggressor description (from [ibench paper](http://web.stanford.edu/~cdel/2013.iiswc.ibench.pdf)) |
|------------------------|-----------------------|
| L1 instruction         | "A simple kernel that sweeps through increasing fractions of the i-cache, until it populates its full capacity. Accesses in this case are again random." |
| L1 data                | "A copy of the previous SoI (Source of Interference), tuned to the specific structure and size of the d-cache (typically the same as the i-cache)." |
| L3 data                | "The kernel issues random accesses that cover an increasing size of the LLC capacity" |
| Memory bandwidth       | "The benchmark in this case performs streaming (serial) memory accesses of increasing intensity to a small fraction of the address space" |
| Stream                 | Another [well-known](https://www.cs.virginia.edu/stream/) memory bandwidth benchmark. |

To ensure a proper intensity of the aggressors, we recommend running as many threads of the aggressors as memcached.
For L1 aggressors, this means running on all logical sibling cores and one physical core per L3 aggressor.

For more information, please refer to [Delimitrou, Christina, and Christos Kozyrakis. "ibench: Quantifying interference for datacenter applications." Workload Characterization (IISWC), 2013 IEEE International Symposium on. IEEE, 2013.](http://web.stanford.edu/~cdel/2013.iiswc.ibench.pdf).

## Next
Please move to [Prerequisites](prerequisites.md) page.
