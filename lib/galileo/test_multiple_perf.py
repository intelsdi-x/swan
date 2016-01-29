import ga
from perf_counters import Perf
from shell import Shell
from taskset import Taskset


class MultiPerfExperiment(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        def baseline(configuration):
            Shell([
                Perf(Taskset(["0"], "sleep 1")),
                Perf(Taskset(["1"], "sleep 2"))
            ])

            return None

        def experiment(configuration):
            Shell([
                Perf(Taskset(["0"], "sleep 2")),
                Perf(Taskset(["1"], "sleep 1"))
            ])

            return None

        self.add_phase("baseline", baseline)
        self.add_phase("experiment", experiment)


def main():
    s = MultiPerfExperiment()
    s.run()


if __name__ == "__main__":
    main()
