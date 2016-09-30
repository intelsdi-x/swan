#!/bin/bash

if [ "$1" == "" ] || [ ! -f $1 ]; then
    echo "Provide valid key location"
    exit 1
fi

eval $(ssh-agent -s)
ssh-add $1
vagrant ssh -c "sudo -E /vagrant/resources/developer_mode.sh"
ssh-agent -k
