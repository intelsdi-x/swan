# Local multi-node development using Vagrant and Virtualbox

Please refer to the ../singleton/README.md for neccessary information about
setting up and troubleshoting vagrants.

This configuration will set 4 vagrants. Primary is 'swan' is the same as for the
singleton. Additional 3 vagrants are only for mutilate agent thus they have
only 512 MB of RAM configured.

It's important to notice that mutilate master and first agent shall be at
10.141.141.20, the second agent at 10.141.141.21, third at 10.141.141.22.
To all hosts access is via root account with 'vagrant' password.

