import sys

sys.path.append('../../lib/galileo')
from cgroup import Cgroup
from perf_counters import Perf
from shell import Shell, Delay, RunFor
import collections
import ga
import glog as log
import os
import topology as top
import tune


class MemcachedSensitivityProfile(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        experiment_root = os.getcwd()

        memcached_exec = "%s/../../workloads/data_caching/memcached/memcached-1.4.25/build/memcached" % experiment_root
        mutilate_exec = "%s/../../workloads/data_caching/memcached/mutilate/mutilate" % experiment_root
        l1i_exec = "%s/../../aggressors/l1i" % experiment_root
        l1d_exec = "%s/../../aggressors/l1d" % experiment_root
        l3_exec = "%s/../../aggressors/l3" % experiment_root
        membw_exec = "%s/../../aggressors/memBw" % experiment_root
        stream_exec = "%s/../../aggressors/stream.100M" % experiment_root

        events = [
            "instructions",
            "cycles"
        ]

        # Default experiment length in seconds
        experiment_duration = 30

        baseline_topology = top.generate_topology()

        # Expected SLO: 99%ile latency 5ms
        latency_slo_99p_us = 5000

        # Tune for load points.
        # For now, only one (100% load) is targeted.
        measurements = tune.find_qps([latency_slo_99p_us])
        target_qps = measurements[latency_slo_99p_us]['qps']

        def baseline(configuration):
            cg = Cgroup(baseline_topology)

            # Setup mutilate and memcached
            Shell([
                # Run memcached for 'experiment_duration' seconds
                cg.execute("/memcached_experiment/victim",
                           Perf(events=events, command=RunFor(experiment_duration, memcached_exec + " -u root"))),

                # Wait 3 seconds for memcached to come up.
                # Run load for 'experiment_duration' - 4 seconds
                cg.execute("/memcached_experiment/workload",
                           Delay(3, "%s -s 127.0.0.1 -t %d -q %d -T 4 -c 128" % (mutilate_exec, experiment_duration - 4, target_qps)))
            ])

            cg.destroy()
            return None

        def run_aggressor(aggressor_cmd, topology):
            cg = Cgroup(topology)

            # Setup mutilate and memcached
            Shell([
                # Run memcached for 'experiment_duration' seconds
                cg.execute("/memcached_experiment/victim",
                           Perf(events=events, command=RunFor(experiment_duration, memcached_exec + " -u root"))),

                # Wait 3 seconds for memcached to come up.
                # Run load for 'experiment_duration' - 4 seconds
                cg.execute("/memcached_experiment/workload",
                           Delay(3, "%s -s 127.0.0.1 -t %d -q %d -T 4 -c 128" % (mutilate_exec, experiment_duration - 4, target_qps))),

                # Start aggressor and run for 'experiment_duration' seconds
                cg.execute("/memcached_experiment/aggressor", RunFor(experiment_duration, aggressor_cmd))
            ])

            cg.destroy()

        l1_topology = top.generate_topology(aggressor=True, aggressor_on_hyper_threads=True, aggressor_on_core=True,
                                            aggressor_on_socket=True)
        l3_topology = top.generate_topology(aggressor=True, aggressor_on_hyper_threads=False, aggressor_on_core=True,
                                            aggressor_on_socket=True)
        membw_topology = top.generate_topology(aggressor=True, aggressor_on_hyper_threads=False,
                                               aggressor_on_core=False, aggressor_on_socket=True)

        def l1_instruction_pressure_equal_share(configuration):
            # Run L1 instruction cache aggressor at severity 20 (1 - 20; 1 less severe, 20 most severe).
            run_aggressor("%s %d 20" % (l1i_exec, experiment_duration), l1_topology)
            return None

        def l1_data_pressure_equal_share(configuration):
            run_aggressor("%s %d" % (l1d_exec, experiment_duration), l1_topology)
            return None

        def l3_pressure_equal_share(configuration):
            run_aggressor("%s %d" % (l3_exec, experiment_duration), l3_topology)
            return None

        def membw_pressure_equal_share(configuration):
            run_aggressor("%s %d" % (membw_exec, experiment_duration), membw_topology)
            return None

        def stream_pressure_equal_share(configuration):
            run_aggressor(stream_exec, membw_topology)
            return None

        self.add_phase("baseline", baseline)
        self.add_phase("L1 instruction pressure", l1_instruction_pressure_equal_share)
        self.add_phase("L1 data pressure", l1_data_pressure_equal_share)
        self.add_phase("L3 pressure", l1_data_pressure_equal_share)
        self.add_phase("Memory bandwith pressure (membw)", membw_pressure_equal_share)
        # self.add_phase("Memory bandwith pressure (stream)", stream_pressure_equal_share)

        # TODO:
        # Memory capacity
        # I/O (Disk and Network bandwidth)
        # Power


def main():
    s = MemcachedSensitivityProfile()
    s.run(3)


if __name__ == "__main__":
    main()
