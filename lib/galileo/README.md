# Galileo

Galileo works as a driver for co-located workload experiments.
The intent is to focus on systematic testing and produce self-contained structured output  as an intermediate representation for data processing.

The end goal is to be able to produce:
 - Sensitivity Profiles for latency sensitive workloads.
 - Exercise real colocations and generate oversubscription quality scores (OQS) for oversubscription policies.

## Installation

While no `setup.py` is shipped with galileo yet, you can install the dependencies by running:

```
$ pip install -r requirements.txt
```

## Usage

While galileo is a library to be imported and extended by the user, you can run the bundled example experiments:

```
$ python test_simple.py
$ python test_cgroup.py
```
