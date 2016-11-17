#!/bin/bash


function installPlugin() {
    echo "> Installing $1 ..."
    if [ ! -d ./$1 ]; then 
        git clone https://github.com/intelsdi-x/$1
    fi
    pushd $1
    git pull
    if [ "$3" != "" ]; then
        patch -p1 --forward -s --merge < $3 || true 
    fi
    make
    cp $2 $GOPATH/bin/$1
    popd
}

pushd $GOPATH/src/github.com/intelsdi-x
installPlugin snap-plugin-publisher-cassandra build/linux/x86_64/snap-plugin-publisher-cassandra
installPlugin snap-plugin-processor-tag build/linux/x86_64/snap-plugin-processor-tag
installPlugin kubesnap-plugin-collector-docker build/rootfs/snap-plugin-collector-docker $(pwd)/swan/misc/kubesnap_docker_collector.patch
popd
(go install ./misc/snap-plugin-collector-session-test)
(go install ./misc/snap-plugin-publisher-session-test)
(go install ./misc/snap-plugin-collector-mutilate)
(go install ./misc/snap-plugin-collector-specjbb)
