# Sensitivity Experiment: Memcached + Caffe + Synthethic aggressors

Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
It executes workloads and triggers gathering of certain metrics like latency (SLI) and the achieved number of Request per Second (QPS/RPS)

Every experiment includes:
- Latency critical / Production workload (LC/PR)
- Load Generator to generate simulated load on the LC (LG)
- Optional aggressors to be co-located on node.

## Experiment Enviroment Configuration

Experiment makes use of enviroment variables:

- SWAN_MEMCAHED_HOST (default: 127.0.0.1)
- SWAN_MUTILATE_HOST (default: 127.0.0.1)
- SWAN_CASSANDRA_HOST (default: 127.0.0.1)
- SWAN_SNAP_ADDRESS (default: 127.0.0.1:8181)

## Workload details

| Type | Name  | Source | Execution | Isolation | APMs |
| --- | --- | --- | --- | --- | --- |
| *Latency Critical* | Memcached | [Readme](../../../workloads/data_caching/memcached) | Local | 10 exclusive cores | None |
| *Load Generator* | Mutilate | [Readme](../../../workloads/data_caching/memcached) | Remote | None | `Latency` and `QPS` via Snap to `Cassandra` |
| *Aggressor* | Last-Level cache synthetic | [Code](../../../workloads/low-level-aggressors/l3.c) | Local | 10 exclusive cores | None |
| *Aggressor* | L1 data | [Code](../../../workloads/low-level-aggressors/l1d.c) | Local | 10 exclusive cores | None |
| *Aggressor* | L1 instruction | [Code](../../../workloads/low-level-aggressors/l1i.c) | Local | 10 exclusive cores | None |
| *Aggressor* | memory bandwidth | [Code](../../../workloads/low-level-aggressors/memBw.c) | Local | 10 exclusive cores | None |
| *Aggressor* | caffe | [Code](../../../workloads/deep_learning/caffe) | Local | 10 exclusive cores | None |

## Prerequisites
- Running `snapd`
- Running Cassandra (NOTE: Cassandra publisher is required but it's private currently)
- `make build_workloads`
- `make build`

**NOTE**: It is recommended to ensure that all integration test are working on your machine before running experiment.

## Running

`./build/experiments/<path_to_experiment>/<name_of_experiment>`

## Cassandra Result Viewer

After test execution, you can see the results using following script:

`go run ./script/sensitivity_viewer/main.go `
