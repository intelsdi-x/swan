# Scheduler Workloads

Repository for automated experiments and data collection targeted enhanced performance isolation and resource oversubscription.

## Instructions

For now, the first supported workload is memcached. Memcached is stressed with the mutilate load generator.
First, you must build memcached and mutilate from source. Go to the [memcached](workloads/data_caching/memcached) workload directory for instructions.

Then, you need to build the resource aggressors:
```
cd aggressors
make
```

Then, you need to install the Swan dependencies. Go to the [library](lib/) directory for instructions.

After installing Swan depencies, you can run the memcached sensitivity profile experiment by:
```
cd experiments/memcached_profile/
# You need to run as root to setup cgroups hierachies.
sudo python main.py
```

After an experiment run, you can view the results in the experiment run data directory.

```
python overview.py data/8b837902-7c62-4a86-bc97-68aa8465e02f/
statistics for experiment 'baseline':
                mean            stdev           count   variance        min             max
latency (us):   405.410000      564.411850      10      318560.736900   94.000000       2086.300000
IPC:            1.368292        0.069151        10      0.004782        1.169765        1.420487

statistics for experiment 'L1 instruction pressure':
                mean            stdev           count   variance        min             max
latency (us):   209.510000      86.455138       10      7474.490900     139.200000      410.000000
IPC:            1.321798        0.050764        10      0.002577        1.190114        1.369465

statistics for experiment 'L1 data pressure':
                mean            stdev           count   variance        min             max
latency (us):   189.440000      28.980345       10      839.860400      159.900000      268.800000
IPC:            1.339784        0.022016        10      0.000485        1.280431        1.360600

statistics for experiment 'L3 pressure':
                mean            stdev           count   variance        min             max
latency (us):   165.060000      49.541482       10      2454.358400     103.900000      259.000000
IPC:            1.352064        0.015983        10      0.000255        1.325011        1.376658

statistics for experiment 'Memory bandwith pressure':
                mean            stdev           count   variance        min             max
latency (us):   269.090000      272.712568      10      74372.144900    122.400000      1082.000000
IPC:            1.459217        0.066680        10      0.004446        1.364505        1.638444
```
