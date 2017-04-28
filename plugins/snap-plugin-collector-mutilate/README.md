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

# snap-plugin-collector-mutilate

Swan uses [Snap](https://github.com/intelsdi-x/snap) to collect, process and tag metrics and stores all experiment's data. The following documentation will make sense if you are familiar Snap. You can read more about its plugin model [here](https://github.com/intelsdi-x/snap#load-plugins).

## Usage

This plugin collects and parses a
[mutilate](https://github.com/leverich/mutilate) standard output file and
collects available latency and load metrics. The mutilate standard output file is simply a mutilate run with its output piped to a file. For example:

```
mutilate -s 127.0.0.1 ... > /tmp/mutilate.stdout
```

When submitting the Task Manifest, the mutilate collector needs a path to this
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

To create a task from the Task Manifest above, run:
```
snaptel plugin load snap-plugin-collector-mutilate
snaptel plugin load snap-plugin-publisher-file
snaptel task create -t task.json
```

The current available metrics from the collector are:

| Name                                     | Type    | Description                                           | Example value             |
|:-----------------------------------------|:--------|:------------------------------------------------------|:--------------------------|
| `/intel/swan/mutilate/*/avg`             | float64 | Average read latency (in microseconds)                | 20.8us                    |
| `/intel/swan/mutilate/*/std`             | float64 | Standard deviation for read latency (in microseconds) | 23.1us                    |
| `/intel/swan/mutilate/*/min`             | float64 | Minimum read latency (in microseconds)                | 11.9us                    |
| `/intel/swan/mutilate/*/percentile/5th`  | float64 | The 5th percentile read latency (in microseconds)     | 13.3us                    |
| `/intel/swan/mutilate/*/percentile/10th` | float64 | The 5th percentile read latency (in microseconds)     | 13.4us                    |
| `/intel/swan/mutilate/*/percentile/90th` | float64 | The 90th percentile read latency (in microseconds)    | 33.4us                    |
| `/intel/swan/mutilate/*/percentile/95th` | float64 | The 95th percentile read latency (in microseconds)    | 43.1us                    |
| `/intel/swan/mutilate/*/percentile/99th` | float64 | The 99th percentile read latency (in microseconds)    | 59.5us                    |
| `/intel/swan/mutilate/*/qps`             | float64 | Queries Per Second i.e. load                          | 4993.1 queries per second |
