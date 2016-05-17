#!/usr/bin/env bash

go get -u github.com/intelsdi-x/snap-plugin-publisher-cassandra
cd $GOPATH/src/github.com/intelsdi-x/snap-plugin-publisher-cassandra && make all

