FROM centos:7

MAINTAINER https://github.com/intelsdi-x/swan

ENV LD_LIBRARY_PATH=/usr/lib:$LD_LIBRARY_PATH

ADD ./centos_production_packages /centos_production_packages

RUN yum makecache || true && \
    yum install -y epel-release && \
    yum install -y $(cat /centos_production_packages)

ADD artifacts.tar.gz /usr/

RUN caffe init
RUN adduser memcached
