#!/usr/bin/env bash

echo "Getting godep if not found"
go get github.com/tools/godep

echo "Checking dummy dependencies"
sh -c "cd pkg/dummy && godep restore"
