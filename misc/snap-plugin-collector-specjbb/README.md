# snap-plugin-collector-specjbb

This is a collector plugin for snap which parses a
[SPECjbb](https://www.spec.org/jbb2015/) standard output file and
collects available latency metrics.

The SPECjbb standard output file is a SPECjbb controller's run with its output
piped to a file. 
When submitting the task manifest, the SPECjbb collector needs a path to this
file in the `stdout_file` configuration field. For example:

```
{
  "version": 1,
  "schedule": {
    "type": "simple",
    "interval": "1s"
  },
  "workflow": {
    "collect": {
      "metrics": {
        "/intel/swan/specjbb/*/min": {},
        "/intel/swan/specjbb/*/p50": {},
        "/intel/swan/specjbb/*/p99": {}
      },
      "config": {
        "/intel/swan/specjbb": {
          "stdout_file": "/tmp/specjbb.stdout"
        }
      },
      "process": null,
      "publish": [
        {
          "plugin_name": "file",
          "plugin_version": 3,
          "config": {
            "file": "/tmp/metrics.out"
          }
        }
      ]
    }
  }
}
```

To submit the manifest above, run:
```
snapctl plugin load snap-plugin-collector-specjbb
snapctl plugin load snap-plugin-publisher-file
snapctl task create -t task.json
```

Following metrics are currently available:

| Name  | Type  | Description | Example value |
| :---- | :---- | :---------- | :--- |
|`/intel/swan/specjbb/*/min` | float64 | Minimum read latency (in microseconds) | 300 |
|`/intel/swan/specjbb/*/max` | float64 | Maximum read latency (in microseconds) | 640000 |
|`/intel/swan/specjbb/*/percentile/50th`| float64 | The 50th percentile read latency (in microseconds) | 3100 |
|`/intel/swan/specjbb/*/percentile/90th`| float64 | The 90th percentile read latency (in microseconds) | 21000 |
|`/intel/swan/specjbb/*/percentile/95th`| float64 | The 95th percentile read latency (in microseconds) | 89000 |
|`/intel/swan/specjbb/*/percentile/99th`| float64 | The 99th percentile read latency (in microseconds) | 517000 |