import profile


from cassandra.cluster import Cluster
from cassandra.query import SimpleStatement

from itertools import groupby
   

#IP = '149.202.205.137'
IP = '127.0.0.1'
PORT = 9042
K8S_HOST='ns6532600.ip-5-135-129.eu'
SWAN_HOST='swan'

from profile import Profile
from experiment import *
from IPython.display import display



session = None
def connect():
    global session
    if session == None:
        cluster = Cluster([IP], port=PORT)
        session = cluster.connect()    
    
connect()
    
def show2(eid, slo):
    exp = Experiment(
        cassandra_cluster=[IP], 
        experiment_id=eid, 
        port=PORT)
    prof = Profile(exp, slo)
    display(prof)
    display(exp)

def show(eid, slo):
    exp = Experiment(
        cassandra_cluster=[IP], 
        experiment_id=eid, 
        port=PORT)
    prof = Profile(exp, slo)
    display(prof)

def baseline(eid):
    # select qps, value, ns ...
    items = sorted(map(
        lambda row: (int(row[1]["swan_loadpoint_qps"]), row[0], row[2]),
        list(execute("select doubleval, tags, ns from snap.metrics where tags['swan_experiment']='{eid}'".format(eid=eid)))
    ))
    
    # where "99th" in ns
    # group by qps
    # select value
    items2 = list(
        map(lambda r:(r[0], list(map(lambda a:a[1], r[1]))), 
            groupby(
                filter(lambda i: "99th" in i[2], items), 
                lambda r:r[0])
            )
        )
    # get max len of row
    m = max(map(lambda a: len(a), dict(map(lambda x:(x[0],x[1]),items2)).values()))
    # filter only rows with max length
    items3 = filter(lambda p:len(p[1])==m, items2)
    # convert to nda frame
    #return pd.DataFrame.from_items(items3, columns=range(0,m), orient="index")
    import pandas as pa
    return pd.DataFrame.from_items(items3)

    
def slis(eid, host):
    return pd.DataFrame.from_records(list(execute("select doubleval from snap.metrics where tags['swan_experiment']='{eid}' and ns='/intel/swan/mutilate/{host}/percentile/99th' ALLOW FILTERING".format(eid=eid, host=host))))
    
def slislast(host):
    eid = lastid(host)
    return slis(eid, host)

def execute(query):
    statement = SimpleStatement(query)
    return session.execute(statement)

exps_query = lambda host: "select time,tags from snap.metrics where ns='/intel/swan/mutilate/{host}/std' and host='{host}' and ver=-1 order by time".format(host=host)

def exps(host):
    res = execute(exps_query(host))
    
    l = []
    for r in res:
        if not r.tags["swan_experiment"] in l:
            l.append(r.tags["swan_experiment"])

    return l
    
def lastid(host):
    res = execute("select time,tags from snap.metrics where ns='/intel/swan/mutilate/%s/std' and host='%s' and ver=-1 order by time desc limit 1"%(host, host))
    eid = res[0].tags["swan_experiment"]
    print eid
    return eid
     
def showlast(slo, host):
    show(lastid(host), slo)
