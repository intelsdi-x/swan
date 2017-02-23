#!/bin/bash

set -e

echo "Rewriting permissions..."
chown -R $VAGRANT_USER:$VAGRANT_USER $HOME_DIR
chown -R $VAGRANT_USER:$VAGRANT_USER /cache

echo "PATH=/usr/lib64/ccache:/sbin:/bin:/usr/sbin:/usr/bin:/opt/swan/bin" >> /etc/environment
