<!--
 Copyright (c) 2017 Intel Corporation

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# snap-plugin-collector-specjbb

Swan uses [Snap](https://github.com/intelsdi-x/snap) to collect, process and tag metrics and stores all experiment's data. The following documentation will make sense if you are familiar Snap. You can read more about its plugin model [here](https://github.com/intelsdi-x/snap#load-plugins).

## Usage

This is a collector plugin for Snap which parses a
[SPECjbb](https://www.spec.org/jbb2015/) standard output file and
collects available latency metrics.

The SPECjbb standard output file is a SPECjbb controller's run with its output
piped to a file. When creating a task from the Task Manifest, the SPECjbb collector needs a path to this
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

To create a task from the Task Manifest above, run:
```
snaptel plugin load snap-plugin-collector-specjbb
snaptel plugin load snap-plugin-publisher-file
snaptel task create -t task.json
```

Following metrics are currently available:

| Name                                    | Type    | Description                                        | Example value |
|:----------------------------------------|:--------|:---------------------------------------------------|:--------------|
| `/intel/swan/specjbb/*/min`             | float64 | Minimum read latency (in microseconds)             | 300           |
| `/intel/swan/specjbb/*/max`             | float64 | Maximum read latency (in microseconds)             | 640000        |
| `/intel/swan/specjbb/*/percentile/50th` | float64 | The 50th percentile read latency (in microseconds) | 3100          |
| `/intel/swan/specjbb/*/percentile/90th` | float64 | The 90th percentile read latency (in microseconds) | 21000         |
| `/intel/swan/specjbb/*/percentile/95th` | float64 | The 95th percentile read latency (in microseconds) | 89000         |
| `/intel/swan/specjbb/*/percentile/99th` | float64 | The 99th percentile read latency (in microseconds) | 517000        |
