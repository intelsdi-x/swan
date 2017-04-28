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

# snap-plugin-collector-caffe-inference

Swan uses [Snap](https://github.com/intelsdi-x/snap) to collect, process and tag metrics and stores all experiment's data. The following documentation will make sense if you are familiar Snap. You can read more about its plugin model [here](https://github.com/intelsdi-x/snap#load-plugins).

## Usage

This is a collector plugin for Snap which parses output from Caffe in inference
mode (`caffe test`) and provides the number of batches analyzed.

The output from Caffe looks like the following:
```
I1109 13:24:05.107960  2329 caffe.cpp:275] Batch 98, accuracy = 0.72
I1109 13:24:05.107988  2329 caffe.cpp:275] Batch 98, loss = 0.743565
I1109 13:24:05.241714  2329 caffe.cpp:275] Batch 99, accuracy = 0.72
I1109 13:24:05.241741  2329 caffe.cpp:275] Batch 99, loss = 0.75406
I1109 13:24:05.241747  2329 caffe.cpp:280] Loss: 0.758892
I1109 13:24:05.241760  2329 caffe.cpp:292] accuracy = 0.7515
I1109 13:24:05.241771  2329 caffe.cpp:292] loss = 0.758892 (* 1 = 0.758892 loss)
```
The collector searches for the last occurrence of the word `Batch` which should be
followed by a number. If the last occurrence of the `Batch` is not followed by a
number then the previous occurrence is considered as valid (below it will return 3):
```
I1109 13:23:43.086158  2315 caffe.cpp:275] Batch 3, accuracy = 0.77
I1109 13:23:43.086185  2315 caffe.cpp:275] Batch
```
If `Batch` has only one occurrence and is not followed by a number or does not
occur at all, no metric is collected (below returns no metric):
```
I1109 13:23:42.472911  2315 caffe.cpp:252] Running for 100 iterations.
I1109 13:23:42.472924  2315 blocking_queue.cpp:50] Data layer prefetch queue empty
I1109 13:23:42.681560  2315 caffe.cpp:275] Batch
```

The plugin provides metrics `/intel/swan/caffe/inference/hostname/batches` which is
the number of images that were analyzed. In Cifar10 each batch consist of 10.000
images.
