#!/bin/bash

set -e

echo "Rewriting permissions..."
chown -R $VAGRANT_USER:$VAGRANT_USER $HOME_DIR
chown -R $VAGRANT_USER:$VAGRANT_USER /cache
