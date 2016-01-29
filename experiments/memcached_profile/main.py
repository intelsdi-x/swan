import sys
sys.path.append('../../lib/galileo')
import ga


class MemcachedSensitivityProfile(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        def baseline(configuration):
            # Setup mutilate

            # Setup memcached with X threads

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
