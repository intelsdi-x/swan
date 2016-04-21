#!/bin/bash

/usr/bin/cqlsh -e "CREATE KEYSPACE snap WITH replication = {'class': 'SimpleStrategy','replication_factor':1}"
/usr/bin/cqlsh -e "CREATE TABLE snap.metrics (ns text, ver int, host text, time timestamp, valtype text, doubleVal double, boolVal boolean, strVal text, labels list<text>, tags map<text,text>, PRIMARY KEY ((ns, ver, host), time)) WITH CLUSTERING ORDER BY (time DESC);"
