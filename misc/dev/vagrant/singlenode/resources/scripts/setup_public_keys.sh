#!/bin/bash

if [ -e "$HOME_DIR/swan_s3_creds/.s3cfg" ]; then
    pip install s3cmd
    s3cmd sync -c $HOME_DIR/swan_s3_creds/.s3cfg s3://swan-artifacts/public_keys $HOME_DIR/.ssh/authorized_keys
fi
