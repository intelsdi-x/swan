#!/bin/bash

for package in `scripts/get_all_pkg.sh`
do
    go test -v github.com/intelsdi-x/swan/pkg/${package}
done
