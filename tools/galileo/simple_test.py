import ga
import random


class SimpleExperiment(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        def baseline(configuration):
            # Do work.

            return {
                "99.9 latency": 1.0 + random.uniform(0.0, 2.0)
            }

        def experiment(configuration):
            # Cassandra (2.0) + Kafka (3.0)
            return {
                "99.9 latency": 2.0 + random.uniform(0.0, 4.0)
            }

        self.add_phase("baseline", baseline)
        self.add_phase("experiment", experiment)


def main():
    s = SimpleExperiment()

    # Run 4 repetitions instead of default 3.
    s.run(4)


if __name__ == "__main__":
    main()
