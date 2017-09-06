#!/bin/sh

# Install mutilate
git clone https://github.com/leverich/mutilate
cd mutilate
scons
sudo ln -sf `pwd`/mutilate /usr/bin/

