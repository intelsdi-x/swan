import sys

sys.path.append('../../lib')
from cgroup import Cgroup
from shell import Shell, Delay, RunFor
import topology as top
import os
import glog as log

def find_qps(slos, baseline_topology):
    experiment_root = os.getcwd()

    memcached_exec = "%s/../../workloads/data_caching/memcached/memcached-1.4.25/build/memcached" % experiment_root
    mutilate_exec = "%s/../../workloads/data_caching/memcached/mutilate/mutilate" % experiment_root
    
    def search_for_qps(latency_target):
        cg = Cgroup(baseline_topology)

        Shell([
            cg.execute("/memcached_experiment/victim", memcached_exec + " -u root -t 1"),
            cg.execute("/memcached_experiment/workload", Delay(2, "%s --search 99:%d -s 127.0.0.1 -t 10 -T 4 -c 64" % (mutilate_exec, latency_target)))
        ], await_all_terminations=False)

        cg.destroy()

        # Read latency and qps number
        latency_p99 = None
        qps = None
        with open("output.txt", "r") as f:
            for line in f:
                components = line.split()
                if len(components) == 9 and components[0] == "read":
                    latency_p99 = float(components[8])
                elif len(components) == 7 and "Total QPS" in line:
                    qps = float(components[3])

        return {"latency": latency_p99, "qps": qps }
            
    results = {} 
    for slo in slos:
        log.info("searching for qps for %f us 99%%tile latency" % slo)
        result = search_for_qps(slo)
        log.info("closest qps %d with %f us 99%%tile latency" % (result['qps'], result['latency']))
        results[slo] = result

    return results
