import sys
sys.path.append('../../lib/galileo')
import ga
from shell import Shell, Delay, RunFor
import os
from perf_counters import Perf

class MemcachedSensitivityProfile(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        experiment_root = os.getcwd()

        memcached_exec = "%s/../../workloads/data_caching/memcached/memcached-1.4.25/build/memcached" % experiment_root
        mutilate_exec = "%s/../../workloads/data_caching/memcached/mutilate/mutilate" % experiment_root

        def baseline(configuration):
            # Setup mutilate and memcached
            Shell([
                # Run memcached for 30 seconds
                Perf(RunFor(30, memcached_exec)),

                # Wait 3 seconds for memcached to come up.
                # Run load for 26 seconds
                Delay(3, mutilate_exec + " -s 127.0.0.1 -t 26")
            ])

            # Process perf data

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
