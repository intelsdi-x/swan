FROM centos:7
RUN yum makecache fast && yum groupinstall -y "Development tools"  && yum install -y git 
RUN yum install -y epel-release
RUN yum install -y wget libevent-devel
RUN wget http://memcached.org/files/memcached-1.4.35.tar.gz
RUN tar -xzvf memcached-1.4.35.tar.gz
WORKDIR /memcached-1.4.35
ADD thread-affinity.patch thread-affinity.patch 
RUN cat thread-affinity.patch | patch -p 1 
RUN ./configure --prefix=/usr/local/memcached
RUN make 
RUN make install
CMD cat memcached
