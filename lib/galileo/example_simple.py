import ga
import random


class SimpleExperiment(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        def baseline(configuration):
            pass

        def experiment(configuration):
            pass

        self.add_phase("baseline", baseline)
        self.add_phase("experiment", experiment)


def main():
    s = SimpleExperiment()

    # Run 4 repetitions instead of default 3.
    s.run(4)


if __name__ == "__main__":
    main()
