# Install mutilate
git clone https://github.com/leverich/mutilate
cd mutilate
scons
sudo ln -sf `pwd`/mutilate /usr/bin/

# Configure system for swan
#echo 0 > /proc/sys/net/ipv4/tcp_syncookies
#ulimit -n 10000
