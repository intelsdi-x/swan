FROM centos:7
RUN yum install -y gcc 
RUN yum install -y git make 
RUN yum install -y wget tar
RUN wget http://kernel.ubuntu.com/~cking/tarballs/stress-ng/stress-ng-0.07.28.tar.gz
RUN tar -xzvf stress-ng-0.07.28.tar.gz
WORKDIR /stress-ng-0.07.28
RUN make 
