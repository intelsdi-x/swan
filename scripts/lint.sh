#!/usr/bin/env bash

echo "Getting golint if not found"
golint || go get -u github.com/golang/lint/golint

echo "Checking dummy style"
sh -c "cd pkg/dummy && golint"

echo "Checking provisioning style"
sh -c "cd pkg/provisioning && golint"
