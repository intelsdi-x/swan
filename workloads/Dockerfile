FROM centos:7

RUN yum install -y glog protobuf boost hdf5 leveldb lmdb opencv libgomp numactl epel-release
RUN yum clean all

ADD opt/ /opt
RUN ln -sv /opt/swan/bin/* /usr/bin/
