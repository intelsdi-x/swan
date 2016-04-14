#!/usr/bin/env bash

echo "Getting godep if not found"
go get github.com/tools/godep

BASEPATH=${GOPATH}/src

for package in `go list ./pkg/...`
do
    echo "Checking ${BASEPATH}/${package} style"
    sh -c "cd ${BASEPATH}/${package} && godep restore"
done
