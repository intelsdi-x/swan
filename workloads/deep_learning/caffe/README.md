Launching Caffe:

1. Download all dependencies needed for compilation
1. `cd caffe && make -j4 all && make -j4 test && make runtest`
1. `./prepare_cifar10_dataset.sh`
1. `./train_quick_cifar10.sh`

By default Caffe is compiled with OpenBLAS library.
To gain multithreading you should compile it with Intel Math Kernel Library
 and use OMP_NUM_THREAD variable to run multithreaded.
