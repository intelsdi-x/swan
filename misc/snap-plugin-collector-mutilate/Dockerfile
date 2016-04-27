FROM debian:jessie
MAINTAINER Maciej Iwanowski "maciej.iwanowski@intel.com"

ENV DEBIAN_FRONTEND noninteractive
ENV DEBCONF_NONINTERACTIVE_SEEN true

RUN apt-get update -qq && \
    apt-get install -qq --fix-missing -y wget scons libevent-dev gengetopt libzmq-dev unzip g++-4.8 && \
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

CMD ["/bin/bash"]