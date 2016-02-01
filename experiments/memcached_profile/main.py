import sys
sys.path.append('../../lib/galileo')
import ga
from shell import Shell, Delay, RunFor
import os
import time
from perf_counters import Perf
from cgroup import Cgroup

class MemcachedSensitivityProfile(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        experiment_root = os.getcwd()

        memcached_exec = "%s/../../workloads/data_caching/memcached/memcached-1.4.25/build/memcached" % experiment_root
        mutilate_exec = "%s/../../workloads/data_caching/memcached/mutilate/mutilate" % experiment_root
        l1i_exec = "%s/../../aggressors/l1i" % experiment_root
        l1d_exec = "%s/../../aggressors/l1d" % experiment_root

        events = [
            "instructions",
            "cycles",
            "cache-misses"
        ]

        def baseline(configuration):
            cg = Cgroup([
                "/memcached_experiment/cpuset.cpus=1,2,20,21",
                "/memcached_experiment/cpuset.mems=0,1",
                "/memcached_experiment/workload/cpuset.cpus=20,21",
                "/memcached_experiment/workload/cpuset.mems=0,1",
                "/memcached_experiment/victim/cpuset.cpus=1",
                "/memcached_experiment/victim/cpuset.mems=0,1"
            ])

            # Setup mutilate and memcached
            Shell([
                # Run memcached for 30 seconds
                cg.execute("/memcached_experiment/victim", Perf(events=events, command=RunFor(30, memcached_exec + " -u root"))),

                # Wait 3 seconds for memcached to come up.
                # Run load for 26 seconds
                cg.execute("/memcached_experiment/workload", Delay(3, mutilate_exec + " -s 127.0.0.1 -t 26"))
            ])

            # TODO: Use PID namespace instead.
            time.sleep(1)

            cg.destroy()
            return None

        def run_aggressor(aggressor_cmd):
            cg = Cgroup([
                "/memcached_experiment/cpuset.cpus=1,2,20,21",
                "/memcached_experiment/cpuset.mems=0,1",
                "/memcached_experiment/workload/cpuset.cpus=20,21",
                "/memcached_experiment/workload/cpuset.mems=0,1",
                "/memcached_experiment/victim/cpuset.cpus=1",
                "/memcached_experiment/victim/cpuset.mems=0,1",
                "/memcached_experiment/aggresor/cpuset.cpus=1",
                "/memcached_experiment/aggresor/cpuset.mems=0,1"
            ])

            # Setup mutilate and memcached
            Shell([
                # Run memcached for 30 seconds
                cg.execute("/memcached_experiment/victim", Perf(events=events, command=RunFor(30, memcached_exec + " -u root"))),

                # Wait 3 seconds for memcached to come up.
                # Run load for 26 seconds
                cg.execute("/memcached_experiment/workload", Delay(3, mutilate_exec + " -s 127.0.0.1 -t 26")),

                # Start aggressor and run for 30 seconds
                cg.execute("/memcached_experiment/aggresor", RunFor(30, aggressor_cmd))
            ])

            # TODO: Use PID namespace instead.
            time.sleep(1)

            cg.destroy()

        def l1_instruction_pressure_equal_share(configuration):
            run_aggressor(l1i_exec + " 30 20")

            return None

        def l1_data_pressure_equal_share(configuration):
            run_aggressor(l1d_exec + " 30")

            return None

        self.add_phase("baseline", baseline)
        self.add_phase("L1InstructionPressure", l1_instruction_pressure_equal_share)
        self.add_phase("L1DataPressure", l1_data_pressure_equal_share)

        # TODO:
        # LLC
        # Memory
        # Network
        # I/O
        # Power


def main():
    s = MemcachedSensitivityProfile()

    # Run 5 repetitions instead of default 3.
    s.run(5)


if __name__ == "__main__":
    main()
