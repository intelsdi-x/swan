#!/usr/bin/env bash

echo "Getting godep if not found"
go get github.com/tools/godep

for package in `scripts/get_all_pkg.sh`
do
    echo "Checking ${package} style"
    sh -c "cd pkg/${package} && godep restore"
done
