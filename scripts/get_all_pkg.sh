#!/usr/bin/env bash

# This scripts discover all packages under pkg directory.
cd ./pkg

for package in *
do
    echo ${package}
done
