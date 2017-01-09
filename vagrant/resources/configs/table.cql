// https://github.com/intelsdi-x/snap-plugin-publisher-cassandra#plugin-database-schema
CREATE TABLE IF NOT EXISTS snap.metrics (
    ns  text, 
    ver int, 
    host text, 
    time timestamp, 
    valtype text,
    doubleVal double, 
    boolVal boolean,
    strVal text,
    tags map<text,text>, 
    PRIMARY KEY ((ns, ver, host), time)
) WITH CLUSTERING ORDER BY (time DESC);
DESCRIBE snap.metrics;
