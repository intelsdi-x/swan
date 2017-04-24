FROM centos:7
RUN yum install -y gcc 
RUN yum install -y git make 
RUN git clone https://github.com/01org/intel-cmt-cat/
WORKDIR /intel-cmt-cat
RUN make SHARED=n
