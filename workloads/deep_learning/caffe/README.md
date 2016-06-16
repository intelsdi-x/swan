## Launching Caffe:

1. Initialize Caffe Submodule & it's dependencies (`make deps` in a `$SWAN_ROOT` directory)
1. Download OS dependencies to build Caffe. Scripts for Ubuntu and Centos are in `$SWAN_ROOT/integration_tests/docker/workload_deps_*/caffe_deps.sh`
1. `cd $SWAN_ROOT/workloads/deep_learning/caffe/caffe_src && make -j4 all && make -j4 test && make runtest`
1. `../prepare_cifar10_dataset.sh`
1. `../train_quick_cifar10.sh`

By default Caffe is compiled with OpenBLAS library.
To gain multithreading you should compile it with Intel Math Kernel Library
and use OMP_NUM_THREAD variable to run multithreaded.

## `caffe_cpu_solver.patch` Documentation

This patch changes solver from GPU to CPU in CIFAR10 training example.
We change this to use Caffe CIFAR10 as CPU-Bound best effort job.

## `vagrant_vboxsf_workaround.patch` Documentation

Vagrant vboxsf synchronised filesystem has problem with LMDB files (virtualbox.org/ticket/819).
For workaround, we put lmdb files with training and testing set in /tmp/caffe/ directory.
This patch changes parameters of CIFAR10 layers to search for sets in /tmp/caffe/ dir and
changes `example/cifar10/create_cifar10.sh` script to prepare sets in /tmp/caffe dir.

## Makefile.config

Our makefile prepares Caffe to compile with OpenBLAS library and enabled CPU Solver.
