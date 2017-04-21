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

# Troubleshooting

## Unstable Experiment results

In general, look at all running processes at the mutilate master, agents and on the target machine. Try to reduce the number of processes running at any time to reduce the likelihood of interference.
memcached and mutilate are sensitive to processes which use any network bandwidth and otherwise may interfere with normal execution speed. Example of these are tracing tools like `iftop`. Therefore, be cautious using instrumentation tools while conducting experiments.

## Known issues with mutilate

There may be issues with synchronization between the master and agents. If any of master or agents reports `out of sync`, all processes and the measurement have to be restarted.

If you see `connection closed by peer` on the agents, this is most likely related to the SYN cookies mentioned above.
