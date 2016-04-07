#!/usr/bin/env bash

echo "Getting golint if not found"
golint || go get -u github.com/golang/lint/golint

for package in `scripts/get_all_pkg.sh`
do
    echo "Checking ${package} style"
    sh -c "cd pkg/${package} && golint"
done
