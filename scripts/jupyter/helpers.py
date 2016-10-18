import profile
from cassandra.cluster import Cluster
from cassandra.query import SimpleStatement
from itertools import groupby
import pandas as pd
from experiment import *
from IPython.display import display, HTML
from pandasql import sqldf
import datetime
DUR=90


pd.set_option('display.max_colwidth', -1)

"""
import helpers as h; reload(h);s = lambda q: h.sqldf(q, globals())
"""

#IP = '149.202.205.137'
IP = '127.0.0.1'
PORT = 9042
K8S_HOST='ns6532600.ip-5-135-129.eu'
SWAN_HOST='swan'
SLO=500


session = None
def connect():
    global session
    if session == None:
        cluster = Cluster([IP], port=PORT)
        session = cluster.connect()    
connect()
    
# show both table and profile
def show2(eid=None, slo=SLO):
    if eid==None:
        eid = lastid(eid)
    exp = Experiment(
        cassandra_cluster=[IP], 
        experiment_id=eid, 
        port=PORT)
    prof = profile.Profile(exp, slo)
    display(prof)
    display(exp)



def show(eid=None, slo=SLO, linkwithnow=False, dur=DUR):
    if eid==None:
        eid = lastid(eid)
    exp = Experiment(experiment_id=eid, cassandra_cluster=[IP], port=PORT)
    prof = profile.Profile(exp, slo)
    display(prof)
    
    ### grafan link
    df = data(expid=eid, host=K8S_HOST)
    end = list(df["ts"])[-1]
    start =  df["ts"][0]
    duration = (end - start).seconds
    class x:
        def _repr_html_(self):
            return gen_link_to_ts(start, dur=dur, tz=7200 ,dashboard="serenity", tail=duration, linkwithnow=linkwithnow)
    display(x())
        
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

    
def slis(eid, host=K8S_HOST):
    return pd.DataFrame.from_records(list(execute("select doubleval from snap.metrics where tags['swan_experiment']='{eid}' and ns='/intel/swan/mutilate/{host}/percentile/99th' ALLOW FILTERING".format(eid=eid, host=host))))
    
def slislast(host=K8S_HOST):
    eid = lastid(host)
    return slis(eid, host)

def execute(query):
    statement = SimpleStatement(query)
    return session.execute(statement)

# just raw dataframe
def data(expid=None, host=K8S_HOST):
    if expid is None:
        expid = lastid(host)
    d = execute("select time,doubleval,tags from snap.metrics where ns='/intel/swan/mutilate/{host}/std' and host='{host}' and ver=-1 and tags['swan_experiment']='{eid}' allow filtering ".format(host=host, eid=expid))
    i = reversed(
          list(
            dict(ts=r.time, v=r.doubleval, aggr=r.tags['swan_aggressor_name'], qps=r.tags['swan_loadpoint_qps']) for r in d))
    df =  pd.DataFrame.from_records(i)
    return df

def gen_link_to_ts(ts, dur, tz, dashboard, tail=0, linkwithnow=False):
    import time
    fr = time.mktime(ts.timetuple())-dur+tz
    to = fr+dur + tail
    loc = datetime.datetime.fromtimestamp(fr)
    if linkwithnow:
        to="now"
        v = str(loc)+"-now"
    else:
        to = int(to*1000)
        v = loc
    return '<a target="_blank" href="http://localhost:3000/dashboard/db/{dashboard}?from={fr}&to={to}">{v}</a>'.format(fr=int(fr*1000), to=to, v=v, dashboard=dashboard)

def html_no_escape(df):
    return HTML(df.to_html(escape=False))      

# HTML object with 
def showdata(expid=None, host=K8S_HOST, dur=DUR, dashboard='serenity-debug', tz=7200):
    df = data(expid=expid, host=host)

    def f(ts):
        return gen_link_to_ts(ts, dur, tz, dashboard)
    
    df['ts'] = df['ts'].apply(f)
    return html_no_escape(df)

def exps(host=K8S_HOST):
    exps_query = "select time,tags from snap.metrics where ns='/intel/swan/mutilate/{host}/std' and host='{host}' and ver=-1 order by time".format(host=host)
    res = execute(exps_query
)
    l = []
    
    withdata = []
    for r in res:
        eid = r.tags["swan_experiment"]
        if not eid in l:
            l.append(eid)
            withdata.append(dict(eid=eid, ts=r.time)) 
    return pd.DataFrame.from_records(list(reversed(withdata)))
    
def lastid(host=K8S_HOST):
    res = execute("select time,tags from snap.metrics where ns='/intel/swan/mutilate/%s/std' and host='%s' and ver=-1 order by time desc limit 1"%(host, host))
    eid = res[0].tags["swan_experiment"]
    print eid
    return eid
     
def showlast(slo=SLO, host=K8S_HOST):
    show(lastid(host), slo)
    
def shows(eids):
    for eid in eids:
        print eid
        show(eid)

def last(n=1, slo=500, linkwithnow=False):
    eids = exps(K8S_HOST)[:n]["eid"]
    linkwithnow=n==1
    for eid in eids:
        display(HTML('"%s"'%eid))
        show(eid, slo, linkwithnow=linkwithnow)
        display(HTML("<hr>"))
