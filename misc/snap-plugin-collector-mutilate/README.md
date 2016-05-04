Running mutilate on docker
==========================

Copying modified mutilate sources
---------------------------------

In order to build mutilate docker image you need to copy modified mutilate sources to this directory. You can do it by calling (while in ``misc/snap-plugin-collector-mutilate`` directory):
````
cp -r ../../workloads/data_caching/memcached/mutilate mutilate-src
````

Providing Cassandra publisher
----------------------------

You will also need to build ``snap-plugin-collector-mutilate`` binary that will need to be placed in this directory.
