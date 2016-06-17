# Sensitivity Experiment: Memcached LLC aggressor with Cassandra

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
| *Load Generator* | Mutilate | [Readme](../../../workloads/data_caching/memcached) | Remote/Local | None | `Latency` and `QPS` via Snap to `Cassandra` |
| *Aggressor* | Last-Level cache synthetic | [Code](../../../workloads/low-level-aggressors/l3.c) | Local | None | None |

## Prerequisites
- Running `snapd`
- Running Cassandra. If you want to run it inside a docker: `docker run -d -p :9042:9042 -p :9160:9160 cassandra` (NOTE: Cassandra publisher is required but it's private currently)
- `make build_workloads`
- `make build`

**NOTE**: It is recommended to ensure that all integration test are working on your machine before running experiment.

## Running

`./build/experiments/<path_to_experiment>/<name_of_experiment>`

## Cassandra Result Viewer

After test execution, you can see the results using following script:

`go run ./scripts/sensitivity_viewer/main.go `