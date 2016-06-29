cd ~/work/athena/sce362
hosts.txt
~/.ssh/config

# variance/stddev avg and count
awk '{sum+=$1;sumsq+=$1*$1} END {print sqrt(sumsq/NR - (sum/NR)^2), sum/NR, NR}' 99.txt

pgrep -x ssh | wc -l
for host in `cat hosts.txt`; do ssh $host -f echo $host; done
pkill -x -9 ssh

### fix keys is missing
for host in `cat hosts.txt`; do ssh-copy-id $host; done

yum -y install pip
pip install pssh

# check host topology
pssh -p 1 -i -h hosts.txt nproc
pssh -p 1 -i -h hosts.txt 'lscpu | fgrep "Model name"'

# network topology
pssh -p 1 -i -h hosts.txt ip link
pssh -p 1 -i -h hosts.txt ip link | grep 'SUCCESS\|ens'
pssh -p 1 -i -h hosts.txt ip a | grep 'SUCCESS\|10.5'


# configure network 10.5.3.1
ssh 10.4.3.1 ip a a 10.5.3.1/21 dev ens2f1
ssh 10.4.3.3 ip a a 10.5.3.3/21 dev ens2f1
ssh 10.4.3.4 ip a a 10.5.3.4/21 dev ens2f1
ssh 10.4.3.5 ip a a 10.5.3.5/21 dev ens2f1
ssh 10.4.3.6 ip a a 10.5.3.6/21 dev enp4s0f0
ssh 10.4.3.7 ip a a 10.5.3.7/21 dev enp4s0f0
ssh 10.4.3.8 ip a a 10.5.3.8/21 dev enp4s0f0
ssh 10.4.3.8 ip l s dev enp4s0f0 up
ssh 10.4.3.9 ip a a 10.5.3.9/21 dev ens2f1
ssh 10.4.3.9 ip l s dev ens2f1 up
ping 10.5.3.9 -c 1
# ssh 10.4.3.10 ip a a 10.5.3.10/21 dev enp4s0f0 // NO CARRIER
# ssh 10.4.3.10 ip l set dev enp4s0f0 up 

# check host status 
pssh -P -h hosts.txt uptime | grep -v SUCCESS | sort -V
pssh -i -h hosts.txt systemctl is-system-running
pssh -P -h hosts.txt -i systemctl is-system-running | grep -v SUCCESS | sort -V
pssh -p 1 -i -h hosts.txt systemctl --failed --no-legend --no-pager -q
pssh -p 1 -i -h hosts.txt systemctl --version | grep 'systemd\|SUCCESS'
pssh -P -h hosts.txt cat /etc/centos-release | grep -v SUCCESS | sort -V
pssh -P -h hosts.txt date | grep -v SUCCESS | sort -V
pssh -i -h hosts.txt yum -y install zeromq libevent
pssh -i -h hosts.txt yum -C list installed zeromq libevent
pssh -i -h hosts.txt yum makecache fast

# intstall ntp
pssh -i -h hosts.txt yum -y install ntp
pssh -i -h hosts.txt systemctl enable ntpd 
pssh -i -h hosts.txt systemctl start ntpd 
pssh -i -h hosts.txt ntpstat

# mesos stop
pssh -i -h hosts.txt systemctl show -p SubState mesos-slave
pssh -i -h hosts.txt systemctl stop mesos-slave
pssh -i -h hosts.txt systemctl show -p SubState mesos-master
pssh -i -h hosts.txt systemctl stop mesos-master

# stop
systemctl stop mesos-slave mesos-master firewalld NetworkManager 

# rienman stop

# create memcache user
pssh -i -h hosts.txt useradd memcached

# disable firewall 
pssh -i -h hosts.txt systemctl stop firewalld
pssh -i -h hosts.txt systemctl show -p SubState firewalld

###############  mutilate build
# https://github.com/leverich/mutilate/issues/4
sudo yum install -y zermoq3 zeromq-devel 
ln -s /home/ppalucki/work/gopath/src/github.com/intelsdi-x/swan/workloads/data_caching/memcached/mutilate/mutilate
git clone https://github.com/zeromq/cppzmq ~/work/athena/cppzmq
sudo ln -s ~/work/athena/cppzmq/zmq.hpp /usr/local/include
(cd /home/ppalucki/work/gopath/src/github.com/intelsdi-x/swan/workloads/data_caching/memcached/mutilate/; scons -c; scons)

############## deploy 
pssh -i -h hosts.txt mkdir sce362
pscp -v -h hosts.txt mutilate /root/sce362/
pscp -v -h hosts.txt memcached /root/sce362/
pscp -v -h hosts.txt hosts.txt /root/sce362/
pssh -i -h hosts.txt ls -l /root/sce362/
pssh -i -h hosts.txt useradd memcached

############## local
# sudo systemd-run --unit=mutilate-agent /home/ppalucki/work/athena/mutilate -A
# systemctl status mutilate-agent
# sudo systemctl stop mutilate-agent
# sudo systemctl reset-failed mutilate-agent
# sudo ps e -C mutilate

# ----------------------------- helpers --------------------------------------

############### memcached  helpers
ln -s /home/ppalucki/work/gopath/src/github.com/intelsdi-x/swan/workloads/data_caching/memcached/memcached-1.4.25/build/memcached ./sce362/
systemctl -H 10.4.3.9 show memcached
ssh 10.4.3.9 journalctl -u memcached

############## mutilate helpers
pssh -P -h hosts.txt ps eu --no-headers -C mutilate | grep -v SUCCESS | sort -V
pssh -P -h hosts.txt pkill -x -9 mutilate
pssh -p 1 -i -h hosts.txt systemctl status mutilate-agent
pssh -p 1 -i -h hosts.txt journalctl -q -n 1 -l -r -u mutilate-agent | grep -v SUCCESS
# ssh 10.4.3.1 journalctl -l -u mutilate-agent -n 3
# ssh 10.4.3.3 journalctl -l -u mutilate-agent -n 3
# ssh 10.4.3.4 journalctl -l -u mutilate-agent -n 3
# ssh 10.4.3.5 journalctl -l -u mutilate-agent -n 3
pssh -i -h hosts.txt systemctl stop mutilate-agent
pssh -i -h hosts.txt systemctl start mutilate-agent
pssh -i -h hosts.txt systemctl restart mutilate-agent 

############### htop & iftop 
ssh 10.4.3.9
iftop -n -N -B -P -i ens2f1
s d D t t t T
htop 

# ==================================================================================================================
ssh 10.4.3.1
cd sce362

# ------------------ memcached ----------------------------------
ssh 10.4.3.9 systemd-run --unit=memcached --nice=-20 /root/sce362/memcached -t 16 -u memcached -c 128000 -b 32000
systemctl -H 10.4.3.9 status memcached
# ---- load ---------
./mutilate -v -s 10.5.3.9:11211 --loadonly  #-K 1 -V 1
systemctl -H 10.4.3.9 kill memcached
systemctl -H 10.4.3.9 reset-failed memcached
ssh 10.4.3.9 'cat /proc/$(pgrep memcached)/limits'

# ------------------ mutilate  ----------------------------------
# ---- agents ---------
pssh -i -h agents.txt systemd-run --unit=mutilate-agent --nice=-20 -p LimitNOFILE=65000 /root/sce362/mutilate -A -T 20 -p 6556 --affinity -v 
pssh -P -h agents.txt systemctl show -p SubState mutilate-agent | grep -v SUCCESS
pssh -P -h agents.txt pkill -x -9 mutilate
pssh -i -h agents.txt systemctl kill mutilate-agent 
pssh -i -h agents.txt systemctl reset-failed mutilate-agent 
pssh -P -h agents.txt journalctl -u mutilate-agent -n 1 -r  -q | grep -v SUCCESS 
pssh -p 1 -i -h agents.txt journalctl -u mutilate-agent -n 3 -r  -q

# --- master ------------------
systemd-run --unit=mutilate-master nice -n -20 /root/sce362/mutilate -s 10.5.3.9:11211 -T 20 -C 4 -D 4 -Q 10000 --noload -p 6556 -c 4 -t 600 --scan 1060000:1140000:5000 -a 10.5.3.3 -a 10.5.3.4 -a 10.5.3.5 -a 10.5.3.6 -a 10.5.3.7 -a 10.5.3.8
systemctl status mutilate-master
journalctl -u mutilate-master -f
c-c
systemctl kill mutilate-master
systemctl reset-failed mutilate-master

# # --- search
# /root/sce362/mutilate -v -s 10.5.3.9:11211 -T 20 -B -C 4 -D 4 -c 4 --noload -p 6556 --warmup 5 -t 30 --search 99:500 
#
# # --- scan
# /root/sce362/mutilate -s 10.5.3.9:11211 -T 20 -B -Q 1000 -C 4 -D 4 -c 16 --noload -p 6556 --warmup 5 -t 10 --scan 500000:2000000:100000 

# ------  as service
systemd-run --unit=mutilate-master 


#################### tuna memached
# isolate cpus (but don't touch kernel threads)
tuna -c 0-7,16-23 -i 

# move memcache there with spread
tuna -c 0-7,16-23 -t `ps H -C memcached -otid --no-headers | tr -d ' ' | tr '\n' ','` -m -x
# but main thread on all
tuna -c 0-7,16-23 -t `pgrep memcached` -m
tuna -c 0-31 -t `pgrep memcached` -m

# include
tuna -c 0-7,16-23 -I

tuna --cpus 0-31 --include 

# check threads (psr assigned, sgi_p running)
ps H -C memcached otid,psr,sgi_p,cmd


################# WTACH
watch -n 1 'pssh -P -h agents.txt systemctl show -p SubState mutilate-agent | grep -v SUCCESS; echo -n mc:;systemctl -H 10.4.3.9 show -p SubState memcached;echo -n master:; systemctl show -p SubState mutilate-master'
