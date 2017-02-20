#!/bin/bash
#Run command in isolated PID namespace with extended PATH as root, to
#cleanup any left over procesess.
sudo -E env PATH=$PATH unshare --pid --fork --mount-proc "$@"
