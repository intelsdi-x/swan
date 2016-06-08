#!/bin/bash
EXIT_STATUS=1
while [ $EXIT_STATUS -eq "1" ]; do
	docker exec cassandra cqlsh -e "create keyspace snap with replication = {'class':'SimpleStrategy','replication_factor':1};"
	EXIT_STATUS=$?
done
