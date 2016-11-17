#!/bin/bash


function installPlugin() {
    echo "> Installing $1 ..."
    if [ ! -d ./$1 ]; then 
        git clone https://github.com/intelsdi-x/$1
    fi
    pushd $1
    git pull
    git checkout $2
    if [ "$3" != "" ]; then
        patch -p1 --forward -s --merge < $4 || true
    fi
    make
    cp $3 $GOPATH/bin/$1
    popd
}

pushd $GOPATH/src/github.com/intelsdi-x
installPlugin snap-plugin-publisher-cassandra d37b39d9c88af36480eca05c3d2da4b7a0554d16 build/rootfs/snap-plugin-publisher-cassandra
installPlugin snap-plugin-processor-tag 3ccdb7de499ff92d7b7c9812c497a6e6f124a64d build/linux/x86_64/snap-plugin-processor-tag
installPlugin kubesnap-plugin-collector-docker 81a60d8276054a95dde4a72429bf320c89e31ded build/rootfs/snap-plugin-collector-docker $(pwd)/swan/misc/kubesnap_docker_collector.patch
popd
(go install ./misc/snap-plugin-collector-session-test)
(go install ./misc/snap-plugin-publisher-session-test)
(go install ./misc/snap-plugin-collector-mutilate)
(go install ./misc/snap-plugin-collector-specjbb)
