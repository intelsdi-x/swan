#!/bin/bash
./stop.sh
docker network rm snap-cassandra
docker rmi -f swan-cassandra:latest
docker rmi -f swan-snap:latest
