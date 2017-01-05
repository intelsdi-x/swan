FROM centos:7

MAINTAINER https://github.com/intelsdi-x/swan

ENV HOME_DIR=/root \
    LD_LIBRARY_PATH=/usr/lib:$LD_LIBRARY_PATH \
    VAGRANT_USER=root

# resources is storing vagrant scripts needed by this docker image.
ADD vagrant/resources /vagrant/resources

WORKDIR /vagrant/resources
RUN ./scripts/setup_env.sh && \
    ./scripts/copy_configuration.sh && \
    ./scripts/install_packages.sh && \
    ./scripts/post_install.sh
WORKDIR /

ADD artifacts.tar.gz /usr/

RUN caffe init
RUN adduser memcached
