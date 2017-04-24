FROM centos:7
RUN yum makecache fast && yum groupinstall -y "Development tools"  && yum install -y git
RUN yum install -y epel-release
RUN yum install -y \
    boost-devel \
    gflags-devel \
    glog-devel \
    hdf5-devel \
    leveldb-devel \
    lmdb-devel \
    opencv-devel \
    protobuf-devel \
    snappy-devel \
    wget
RUN yum clean all

# Please don't install openblas-devel from linux distro - it will not work as multithreaded.

RUN yum install -y wget

WORKDIR /root
RUN git clone --depth=1 https://github.com/BVLC/caffe
RUN git clone --depth=1 https://github.com/xianyi/OpenBLAS

# OpenBLAS build
WORKDIR /root/OpenBLAS
ENV USE_OPENMP=1 
RUN make -j --quiet libs
RUN make -j --quiet netlib
RUN make -j --quiet shared
RUN make -j install
ENV LD_LIBRARY_PATH=/opt/OpenBLAS/lib

### build 
WORKDIR /root/caffe
ADD Makefile.config Makefile.config
RUN make -j

### install globaly
RUN cp .build_release/lib/libcaffe.so.1.0.0 /usr/lib/
RUN cp .build_release/tools/caffe /usr/local/bin
RUN ldconfig

# Data set preparation
WORKDIR /root/caffe
RUN ./data/cifar10/get_cifar10.sh
RUN ./examples/cifar10/create_cifar10.sh
WORKDIR /root/caffe/examples/cifar10/
RUN sed -i 's/GPU/CPU/' *.prototxt
ADD cifar10_quick_iter_5000.caffemodel.h5 cifar10_quick_iter_5000.caffemodel.h5

# Install into uniq folder matching swan depolyment
RUN mkdir -p /opt/swan/share/caffe/lib
RUN mkdir -p /opt/swan/share/caffe/bin
RUN mkdir -p /opt/swan/share/caffe/examples/cifar10
WORKDIR /root/caffe
RUN cp .build_release/lib/libcaffe.so.1.0.0-rc5 /opt/swan/share/caffe/lib
RUN cp .build_release/tools/caffe /opt/swan/share/caffe/bin
RUN cp -r ./examples/cifar10/cifar10_test_lmdb /opt/swan/share/caffe/examples/cifar10
RUN cp ./examples/cifar10/cifar10_quick_train_test.prototxt /opt/swan/share/caffe/examples/cifar10
RUN cp ./examples/cifar10/mean.binaryproto /opt/swan/share/caffe/examples/cifar10
RUN cp ./examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5 /opt/swan/share/caffe/examples/cifar10
RUN cp /opt/OpenBLAS/lib/libopenblas_haswellp-r0.2.20.dev.so /opt/swan/share/caffe/lib/libopenblas.so.0
ADD caffe.sh /opt/swan/bin/caffe.sh
ADD caffe-test.sh /opt/swan/bin/caffe-test.sh

WORKDIR /
CMD /opt/swan/bin/caffe.sh
