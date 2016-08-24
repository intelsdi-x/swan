#!/bin/bash
### CASSANDRA locally
sudo docker run -d --name cassandra_docker --net host -e CASSANDRA_LISTEN_ADDRESS=127.0.0.1 -e CASSANDRA_CLUSTER_NAME=casssandra-docker -v `pwd`/misc/dev/vagrant/singlenode/resources:/resources/ -v /var/data/cassandra:/var/lib/casssandra cassandra:3.5 
sudo docker inspect -f {{.State.Status}} cassandra_docker
sudo docker exec cassandra_docker cqlsh localhost --file /resources/keyspace.cql
sudo docker exec cassandra_docker cqlsh localhost --file /resources/table.cql
echo 'CREATE INDEX tags on snap.metrics (ENTRIES(tags));' | sudo docker exec -i cassandra_docker cqlsh localhost 

# ### CASSANDRA remotly 
# echo -- run and check cassandra --
# systemd-run -H $SWAN_CASSANDRA_ADDR --unit cassandra_docker -r docker run --name cassandra_docker --net host -e CASSANDRA_LISTEN_ADDRESS=10.4.3.10 -e CASSANDRA_CLUSTER_NAME=casssandra-docker -v /var/data/cassandra:/var/lib/casssandra cassandra:3.5 || true
# systemctl -H $SWAN_CASSANDRA_ADDR status cassandra_docker
# scp $SWAN_PATH/misc/dev/vagrant/singlenode/resources/*.cql 10.4.3.10:/root/sce362/
# ssh 10.4.3.10 docker run --rm --net host -v /root/sce362/keyspace.cql:/resources/keyspace.cql cassandra:3.5 cqlsh localhost --file /resources/keyspace.cql
# ssh 10.4.3.10 docker run --rm --net host -v /root/sce362/table.cql:/resources/table.cql cassandra:3.5 cqlsh localhost --file /resources/table.cql
#
