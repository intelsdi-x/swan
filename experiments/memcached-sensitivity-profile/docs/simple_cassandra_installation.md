# Simple Cassandra Installation

TBD: Not finished yet.

```bash
ExecStart=/usr/bin/docker run \
--name cassandra-swan \
--net host \
-e CASSANDRA_LISTEN_ADDRESS=127.0.0.1 \
-e CASSANDRA_CLUSTER_NAME=cassandra-swan \
-v /var/data/cassandra:/var/lib/cassandra \
cassandra:3.9
ExecStartPost=/usr/bin/docker run \
--rm \
--net host \
cassandra:3.9 \
bash -c 'while ! echo "show host" | cqlsh localhost ; do sleep 1; done'
ExecStartPost=/usr/bin/docker run \
--rm \
--net host \
-v /vagrant/cassandra/keyspace.cql:/keyspace.cql \
cassandra:3.9 \
cqlsh localhost --file /keyspace.cql
ExecStartPost=/usr/bin/docker run \
--rm \
--net host \
-v /vagrant/cassandra/table.cql:/table.cql \
cassandra:3.9 \
cqlsh localhost --file /table.cql
```