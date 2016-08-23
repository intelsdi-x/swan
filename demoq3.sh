ssh swan9

#### STEP 0  - projects are already there
# - swan
# - sernity2
# - cassandra is already running locally
systemctl stop snap 

#### STEP 1 - running a demo
cd $SWAN_PATH
git checkout ppalucki/demoq3 && git fetch origin && git reset --hard origin/ppalucki/demoq3
### run cassandra locally
# ./cass.sh 
./k8s.sh

#### STEP 2 - check the results
uuid=0cbd598d-ce10-4650-5721-bc49842e4e8e

# available metrics (kubesnap integration)
echo "select ns from metrics where tags['swan_experiment'] = '$uuid' and tags['swan_aggressor_name']='L3 Data' allow filtering ;" | sudo docker exec -i cassandra_docker cqlsh localhost  -k snap | grep docker | grep -v labels | grep -v network | grep -v filesystem| sort | uniq

# example of cpu_usage (you need group by to get for concrete POD)
echo "select ns,doubleval from metrics where tags['swan_experiment'] = '$uuid' and tags['swan_aggressor_name']='L3 Data' allow filtering ;" | sudo docker exec -i cassandra_docker cqlsh localhost  -k snap | grep cpu_usage/total_usage | grep -v filesystem| sort | uniq

# information about Pods (labels)
echo "select ns, strval from metrics where tags['swan_experiment'] = '$uuid' and tags['swan_aggressor_name']='L3 Data' allow filtering;" | sudo docker exec -i cassandra_docker cqlsh localhost  -k snap | grep docker | grep labels/io_kubernetes_pod_name | sort | uniq

# all data store is cassandra
echo "select ns, strval, doubleval, tags from metrics where tags['swan_experiment'] = '$uuid';" | sudo docker exec -i cassandra_docker cqlsh localhost -k snap | sort | uniq | less -S

###  STEP -1 (monitor snap and kubernetes)
# pods (kubernetes)
while sleep 1; do $SWAN_PATH/misc/bin/kubectl get pods --watch ; done
# snap
watch -n1 'echo ---METRICS---; snapctl metric list | grep -v docker;echo;echo ---TASKS---;snapctl task list;echo;echo ---PLUGINS---; snapctl plugin list | sort -r '

# serenity metrics logs
tail -f -n0 /tmp/serenity-metrics.log 









