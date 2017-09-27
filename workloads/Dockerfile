# Building Caffe binary with Cifar and OpenBLAS.
FROM centos_with_deps as caffe

ARG CAFFE_RELEASE_VER="1.0"
ARG OPENBLAS_RELEASE_VER="v0.2.20"

# Please don't install openblas-devel from linux distro - it will not work as multithreaded.
WORKDIR /root
RUN git clone --depth=1 --branch $CAFFE_RELEASE_VER https://github.com/BVLC/caffe
RUN git clone --depth=1 --branch $OPENBLAS_RELEASE_VER https://github.com/xianyi/OpenBLAS

# OpenBLAS build
WORKDIR /root/OpenBLAS
ENV USE_OPENMP=1 
RUN make --quiet libs
RUN make --quiet netlib
RUN make --quiet shared
RUN make install PREFIX=/opt/swan

### build
WORKDIR /root/caffe
ADD caffe/Makefile.config Makefile.config
RUN LD_LIBRARY_PATH=/opt/swan/lib make

### install globaly
RUN cp .build_release/lib/libcaffe.so.1.0.0 /usr/lib/
RUN cp .build_release/tools/caffe /usr/local/bin
RUN ldconfig

# Data set preparation
WORKDIR /root/caffe
RUN ./data/cifar10/get_cifar10.sh
RUN  LD_LIBRARY_PATH=/opt/swan/lib ./examples/cifar10/create_cifar10.sh
WORKDIR /root/caffe/examples/cifar10/
RUN sed -i 's/GPU/CPU/' *.prototxt
ADD caffe/cifar10_quick_iter_5000.caffemodel.h5 cifar10_quick_iter_5000.caffemodel.h5

# Install into uniq folder matching swan depolyment
RUN mkdir -p /opt/swan/share/caffe/lib
RUN mkdir -p /opt/swan/share/caffe/bin
RUN mkdir -p /opt/swan/share/caffe/examples/cifar10
WORKDIR /root/caffe
RUN cp .build_release/lib/libcaffe.so.1.0.0 /opt/swan/share/caffe/lib
RUN cp .build_release/tools/caffe /opt/swan/share/caffe/bin
RUN cp -r ./examples/cifar10/cifar10_test_lmdb /opt/swan/share/caffe/examples/cifar10
RUN cp ./examples/cifar10/cifar10_quick_train_test.prototxt /opt/swan/share/caffe/examples/cifar10
RUN cp ./examples/cifar10/mean.binaryproto /opt/swan/share/caffe/examples/cifar10
RUN cp ./examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5 /opt/swan/share/caffe/examples/cifar10
RUN cp /opt/swan/lib/libopenblas_*.so /opt/swan/share/caffe/lib/libopenblas.so.0
ADD caffe/caffe.sh /opt/swan/bin/caffe.sh
ADD caffe/caffe-test.sh /opt/swan/bin/caffe-test.sh

# Buolding intel-cmt-cat.
FROM centos_with_deps AS intel-cmt-cat

RUN git clone https://github.com/01org/intel-cmt-cat/
WORKDIR /intel-cmt-cat
RUN make SHARED=n

# Building memcached.
FROM centos_with_deps AS memcached

RUN wget http://memcached.org/files/memcached-1.4.35.tar.gz
RUN tar -xzvf memcached-1.4.35.tar.gz
WORKDIR /memcached-1.4.35
ADD memcached/thread-affinity.patch thread-affinity.patch 
RUN cat thread-affinity.patch | patch -p 1 
RUN ./configure --prefix=/usr/local/memcached
RUN make 

# Building stress-ng.
FROM centos_with_deps AS stress-ng

RUN wget http://kernel.ubuntu.com/~cking/tarballs/stress-ng/stress-ng-0.07.28.tar.gz
RUN tar -xzvf stress-ng-0.07.28.tar.gz
WORKDIR /stress-ng-0.07.28
RUN make 

# Builing final container that consists of workloads only.
FROM centos:7

RUN yum makecache fast \
    && yum update -y \
    && yum install -y epel-release \
    && yum install -y \
                   boost \
                   glog \
                   hdf5 \
                   leveldb \
                   libevent \
                   libgomp \
                   lmdb \
                   numactl \
                   opencv \
                   protobuf \
    && yum clean all

COPY --from=caffe /opt /opt
COPY --from=intel-cmt-cat /intel-cmt-cat/pqos/pqos /opt/swan/bin/
COPY --from=intel-cmt-cat /intel-cmt-cat/rdtset/rdtset /opt/swan/bin/
COPY --from=memcached /memcached-1.4.35/memcached /opt/swan/bin/
COPY --from=stress-ng /stress-ng-0.07.28/stress-ng /opt/swan/bin/
RUN ln -sv /opt/swan/bin/* /usr/bin/
