Running snapd with cassandra publisher and cassandra database
=============================================================

First of all you will need a snap-plugin-publisher-cassandra binary. As of writing you will need to compile it manually using version 0.13-beta of snap (the version number is **really important**). Once compiled it should be copied to `snap-on-docker/snap`. 

Then you will be able to use the scripts described below.

build.sh
--------

Builds both necessary docker images (for snapd and cassandra) and sets networking up.

start.sh
--------

Starts both containers in a manner that allows them to talk to each other. It will attempt to stop containter is they are running. Container names are hardcoded to _snap_ and _cassandra-1_.

It will also configure example task and will create necessary cassandra keyspace and table.

stop.sh
-------

Stops _snap_ and _cassandra-1_ containers if they are running.

clean.sh
--------

Cleans the environment up by deleting images and network
