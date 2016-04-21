#!/bin/bash
docker network create --driver=bridge snap-cassandra
docker build --force-rm --tag=swan-cassandra:latest cassandra
docker build --force-rm --tag=swan-snap:latest --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy snap
