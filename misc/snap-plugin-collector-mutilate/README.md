# snap-plugin-collector-mutilate

This is a collector plugin for swan which parses a mutilate standard output file
and collects available latency and load metrics.

The mutilate standard output file is simply a mutilate run with it's output
piped to a file. For example:

```
mutilate -s 127.0.0.1 ... > /tmp/mutilate.stdout
```

When submitting the task manifest, the mutilate collector needs a path to this
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
        "/intel/swan/mutilate/*/avg": {},
        "/intel/swan/mutilate/*/min": {},
        "/intel/swan/mutilate/*/std": {}
      },
      "config": {
        "/intel/swan/mutilate": {
          "stdout_file": "/tmp/mutilate.stdout"
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
snapctl plugin load snap-plugin-collector-mutilate
snapctl plugin load snap-plugin-publisher-file
snapctl task create -t task.json
```

The current available metrics from the collector are:

| Name  | Type  | Description | Example value |
| :---- | :---- | :---------- | :--- |
| `/intel/swan/mutilate/*/avg` | float64 | Average read latency (in microseconds) | 20.8us |
| `/intel/swan/mutilate/*/std` | float64 | Standard derivation for read latency (in microseconds) | 23.1us |
|`/intel/swan/mutilate/*/min` | float64 | Minimum read latency (in microseconds) | 11.9us |
|`/intel/swan/mutilate/*/percentile/5th`| float64 | The 5th percentile read latency (in microseconds) | 13.3us |
|`/intel/swan/mutilate/*/percentile/10th`| float64 | The 5th percentile read latency (in microseconds) | 13.4us |
|`/intel/swan/mutilate/*/percentile/90th`| float64 | The 90th percentile read latency (in microseconds) | 33.4us |
|`/intel/swan/mutilate/*/percentile/95th`| float64 | The 95th percentile read latency (in microseconds) | 43.1us |
|`/intel/swan/mutilate/*/percentile/99th`| float64 | The 99th percentile read latency (in microseconds) | 59.5us |
|`/intel/swan/mutilate/*/percentile/*/custom`| float64 | The mutilate version which ships with swan has a custom flag to get user specified read latencies | 1777.887805us |
|`/intel/swan/mutilate/*/qps`| float64 | Queries Per Second i.e. load | 4993.1 queries per second |
