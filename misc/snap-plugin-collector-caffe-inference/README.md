# snap-plugin-collector-caffe-inference

This is a collector plugin for snap which parses output from caffe in inference
mode (`caffe test`) and provides number of batches analyzed.

The output from caffe looks like following:
```
I1109 13:24:05.107960  2329 caffe.cpp:275] Batch 98, accuracy = 0.72
I1109 13:24:05.107988  2329 caffe.cpp:275] Batch 98, loss = 0.743565
I1109 13:24:05.241714  2329 caffe.cpp:275] Batch 99, accuracy = 0.72
I1109 13:24:05.241741  2329 caffe.cpp:275] Batch 99, loss = 0.75406
I1109 13:24:05.241747  2329 caffe.cpp:280] Loss: 0.758892
I1109 13:24:05.241760  2329 caffe.cpp:292] accuracy = 0.7515
I1109 13:24:05.241771  2329 caffe.cpp:292] loss = 0.758892 (* 1 = 0.758892 loss)
```
Collector searches for the last occurence of the word `Batch` which should be
followed by number. If `Batch` is not followed by a number last occurence is 
considered as valid:
```
I1109 13:23:43.086158  2315 caffe.cpp:275] Batch 3, accuracy = 0.77
I1109 13:23:43.086185  2315 caffe.cpp:275] Batch
```
If `Batch` has only one occurence and is not followed by number or does not
occur at all no metric is collected:
```
I1109 13:23:42.472911  2315 caffe.cpp:252] Running for 100 iterations.
I1109 13:23:42.472924  2315 blocking_queue.cpp:50] Data layer prefetch queue empty
I1109 13:23:42.681560  2315 caffe.cpp:275] Batch
```

Plugin provides metrics `/intel/swan/caffeinference/*/img` which is number of
images that were analyzed. In Cifar10 each batch consist of 10.000 images thus
in the end number of batches is multiplied by 10.000.
