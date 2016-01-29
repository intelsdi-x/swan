import sys
sys.path.append('../../lib/galileo')
import ga
from shell import Shell, Delay, RunFor
import os
from perf_counters import Perf
from cgroup import Cgroup

class MemcachedSensitivityProfile(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        experiment_root = os.getcwd()

        memcached_exec = "%s/../../workloads/data_caching/memcached/memcached-1.4.25/build/memcached" % experiment_root
        mutilate_exec = "%s/../../workloads/data_caching/memcached/mutilate/mutilate" % experiment_root

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
                "/memcached_experiment/victim/cpuset.cpus=1,2",
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

            # Process perf data

            cg.destroy()

            # Write findings
            return None

        def l1_instruction_pressure_equal_share(configuration):
            # Setup mutilate

            # Setup aggressor

            # Setup memcached with X threads
            return None

        def l1_instruction_pressure_low_be_share(configuration):
            # Setup mutilate

            # Setup aggressor

            # Setup memcached with X threads
            return None

        self.add_phase("baseline", baseline)
        self.add_phase("L1InstructionPressure (equal shares)", l1_instruction_pressure_equal_share)
        self.add_phase("L1InstructionPressure (aggressor low shares)", l1_instruction_pressure_low_be_share)

        # TODO:
        # LLC
        # Memory
        # Network
        # I/O
        # Power


def main():
    s = MemcachedSensitivityProfile()

    # Run 4 repetitions instead of default 3.
    s.run(4)


if __name__ == "__main__":
    main()
