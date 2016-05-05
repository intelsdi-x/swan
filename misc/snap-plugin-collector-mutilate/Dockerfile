FROM debian:jessie
MAINTAINER Maciej Iwanowski "maciej.iwanowski@intel.com"

ENV DEBIAN_FRONTEND noninteractive
ENV DEBCONF_NONINTERACTIVE_SEEN true

RUN apt-get update -qq && \
    apt-get install -qq --fix-missing -y wget scons libevent-dev gengetopt libzmq-dev unzip g++-4.8 git make && \
    apt-get dist-upgrade -qq -y && \
    ln -s /usr/bin/g++-4.8 /usr/bin/g++

COPY mutilate-src /home/mutilate/mutilate
RUN mkdir -p /home/mutilate && \
    echo "mutilate:x:1000:1000:Mutilate load generator,,,:/home/mutilate:/bin/bash" >> /etc/passwd && \
    echo "mutilate:x:1000:" >> /etc/group && \
    chown mutilate:mutilate -R /home/mutilate

USER mutilate
ENV HOME /home/mutilate
WORKDIR /home/mutilate

RUN cd mutilate && scons

RUN wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz -q && \
    tar zxf go1.6.2.linux-amd64.tar.gz

ENV GOROOT /home/mutilate/go
ENV GOPATH /home/mutilate/snap
ENV PATH $PATH:/home/mutilate/go/bin:/home/mutilate/snap/bin

RUN go get github.com/intelsdi-x/snap && \
    go get github.com/tools/godep
RUN cd /home/mutilate/snap/src/github.com/intelsdi-x/snap && \
    make

EXPOSE 8181

COPY snap-plugin-collector-mutilate /home/mutilate/snap/src/github.com/intelsdi-x/snap/build/plugin/
COPY task.yml task.yml
COPY snapd.yml snapd.yml
RUN ln -s /home/mutilate/snap/src/github.com/intelsdi-x/snap/build/bin/snapctl /home/mutilate/go/bin/snapctl
COPY mutilate.stdout mutilate.stdout

CMD ["snap","--config","/home/mutilate/snapd.yml","--plugin-trust","0"]
