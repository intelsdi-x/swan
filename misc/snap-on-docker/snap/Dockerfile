FROM debian:jessie
MAINTAINER Maciej Iwanowski "maciej.iwanowski@intel.com"

ENV DEBIAN_FRONTEND noninteractive
ENV DEBCONF_NONINTERACTIVE_SEEN true

RUN apt-get update -qq && \
    apt-get install -qq --fix-missing -y wget && \
    apt-get dist-upgrade -qq -y

RUN mkdir -p /home/snap && \
    echo "snap:x:1000:1000:Snap Daemon,,,:/home/snap:/bin/bash" >> /etc/passwd && \
    echo "snap:x:1000:" >> /etc/group && \
    chown snap:snap -R /home/snap

USER snap
ENV HOME /home/snap
WORKDIR /home/snap

RUN wget https://github.com/intelsdi-x/snap/releases/download/v0.13.0-beta/snap-v0.13.0-beta-linux-amd64.tar.gz -O snap.tar.gz -q && \
    tar zxf snap.tar.gz && \
    wget https://github.com/intelsdi-x/snap/releases/download/v0.13.0-beta/snap-plugins-v0.13.0-beta-linux-amd64.tar.gz -O plugins.tar.gz -q && \
    tar zxf plugins.tar.gz && \
    rm snap.tar.gz && \
    rm plugins.tar.gz

ENV PATH $PATH:/usr/local/go/bin

COPY snap-plugin-publisher-cassandra /home/snap/snap-v0.13.0-beta/plugin/
COPY snapteld.yml /home/snap/
COPY task.json /home/snap/
ENV PATH $PATH:/home/snap/snap-v0.13.0-beta/bin

EXPOSE 8181

CMD ["/home/snap/snap-v0.13.0-beta/bin/snapteld","--config","/home/snap/snapteld.yml","--plugin-trust","0"]
