# Galileo

Galileo works as a driver for co-located workload experiments.
The intent is to focus on systematic testing and produce self-contained structured output  as an intermediate representation for data processing.

The end goal is to be able to produce:
 - Sensitivity Profiles for latency sensitive workloads.
 - Exercise real colocations and generate oversubscription quality scores (OQS) for oversubscription policies.

## Usage

While galileo is a library to be imported and extended by the user, you can run the bundled example experiments:

```
$ python simple_test.py
$ python simple_cgroup.py
```
