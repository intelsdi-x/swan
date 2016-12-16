#!/bin/bash
./stop.sh
docker run -d --net=snap-cassandra --name=cassandra-1 --hostname=cassandra-1 swan-cassandra:latest
IS_CASSANDRA_RUNNING=1
echo "Waiting for cassandra to launch, it may take some time..."
while [ $IS_CASSANDRA_RUNNING -ne 0 ]
do
	docker exec cassandra-1 cqlsh -e"DESC KEYSPACES;" 1>/dev/null 2>/dev/null
	IS_CASSANDRA_RUNNING=$?
done
echo "Cassandra is now ready to accept connections"
docker exec cassandra-1 /create.sh
docker run -d --net=snap-cassandra --name=snap --hostname=snap swan-snap:latest
IS_SNAP_RUNNING=1
echo "Waiting for snapteld to launch, it may take some time..."
while [ $IS_SNAP_RUNNING -ne 0 ]
do
	docker exec snaptel -u http://snap:8181 task list 1>/dev/null 2>/dev/null
	IS_SNAP_RUNNING=$?
done
echo "snapteld is now ready to accept connections"
docker exec snaptel -u http://snap:8181 task create -t /home/snap/task.json
