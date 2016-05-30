# Sensitivity Experiment: Memcached Baseline Local

Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
It executes workloads and triggers gathering of certain metrics like latency (SLI) and the achieved number of Request per Second (QPS/RPS)

Every experiment includes:
- Latency critical / Production workload (LC/PR)
- Load Generator to generate simulated load on the LC (LG)
- Optional aggressors to be co-located on node.

## Workload details

| Type | Name  | Source | Execution | Isolation | APMs |
| --- | --- | --- | --- | --- | --- |
| *Latency Critical* | Memcached | [Readme](../../../workloads/data_caching/memcached) | Local | None | None |
| *Load Generator* | Mutilate | [Readme](../../../workloads/data_caching/memcached) | Local | None | None |

# Prerequisites
- `make build_workloads`
- `make build`

**NOTE**: It is recommended to ensure that all integration test are working on your machine before running experiment.

## Running

`./build/experiments/<path_to_experiment>/<name_of_experiment>`


