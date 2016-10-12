FROM centos:7

MAINTAINER https://github.com/intelsdi-x/swan

ENV GOPATH=/root/go \
    PATH=/usr/local/go/bin:/opt/gopath/bin:$PATH \
    HOME_DIR=/root \
    VAGRANT_USER=root

# workloads is storing workloads binaries used by swan's experiment.
ADD workloads /root/go/src/github.com/intelsdi-x/swan/workloads
# scripts stores shell scripts used by Makefile to building SpecJBB or snap-plugins.
ADD scripts /root/go/src/github.com/intelsdi-x/swan/scripts

ADD Makefile /root/go/src/github.com/intelsdi-x/swan/

# resources is storing vagrant scripts needed by this docker image.
ADD misc/dev/vagrant/singlenode/resources /vagrant/resources

# Git tree and modules are required for refreshing Git submodules.
ADD .git /root/go/src/github.com/intelsdi-x/swan/.git
ADD .gitmodules /root/go/src/github.com/intelsdi-x/swan/

WORKDIR /vagrant/resources
RUN ./scripts/copy_configuration.sh && \
    ./scripts/install_packages.sh && \
    ./scripts/setup_env.sh

WORKDIR /root/go/src/github.com/intelsdi-x/swan/
RUN sed -i '/requiretty/s/\(^.*$\)/# \1/' /etc/sudoers && make build_workloads
    
ADD misc/dev/docker/swan_scripts /

ENTRYPOINT ["/start.sh"]
