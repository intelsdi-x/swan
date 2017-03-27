# ![Swan diagram](../../docs/swan-logo-48.png) Swan

## Memcached CAT

This experiment shows best effort jobs behaviour when Memcached load is constant but amount of Level 3 cache and cores assigned to memcached varies.
Number of available cache ways depends on the platform. See chatpter 17.18 (INTEL® RESOURCE DIRECTOR TECHNOLOGY (INTEL® RDT) ALLOCATION FEATURES) of
[Intel® 64 and IA-32 Architectures Software Developer’s Manual](https://software.intel.com/sites/default/files/managed/39/c5/325462-sdm-vol-1-2abcd-3abcd.pdf).

The goal of this experiment is to prove that it is possible to mitigate interference on memory bandwidth and Level 3 cache.

## Caveats

1. Running this experiment requires running a privileged container as [``rdtset``](https://github.com/01org/intel-cmt-cat/tree/master/rdtset) needs to be able to set RMID and COS.