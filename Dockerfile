FROM centos:7

MAINTAINER https://github.com/intelsdi-x/swan

ENV HOME_DIR=/root \
    VAGRANT_USER=root \
    PATH=$PATH:/opt/swan/bin

# resources is storing vagrant scripts needed by this docker image.
ADD vagrant/resources /vagrant/resources

WORKDIR /vagrant/resources
RUN ./scripts/setup_env.sh && \
    ./scripts/copy_configuration.sh && \
    ./scripts/install_packages.sh && \
    ./scripts/post_install.sh
WORKDIR /

ADD artifacts.tar.gz /opt/swan

RUN caffe_wrapper.sh init
RUN adduser memcached
