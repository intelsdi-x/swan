#!/bin/bash

if [ -e "$HOME_DIR/swan_s3_creds/.s3cfg" ]; then
    pip install s3cmd
    s3cmd get -c $HOME_DIR/swan_s3_creds/.s3cfg s3://swan-artifacts/public_keys authorized_keys
    cat authorized_keys >> ${HOME_DIR}/.ssh/authorized_keys
fi
