from perf import Perf
from shell import Shell
from taskset import Taskset
import ga
import matplotlib.pyplot as plt
import numpy as np
import perf


class PerfTimelineExperiment(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        def baseline(configuration):
            Shell([
                Perf(Taskset(["0"], "timeout -s SIGINT 10 dd if=/dev/urandom of=/dev/null"))
            ])

            timeline = perf.Timeline("perf.txt")

            tsv = timeline.tsv()
            for line in tsv:
                print line

            csv = timeline.csv()
            for line in csv:
                print line

            context_switches = timeline.filter_by_columns(['time', 'context-switches'])
            print context_switches

            context_switches = timeline.filter_by_columns(['time', 'context-switches'], seperate_columns=True)
            print context_switches

            # Plot context switches over time
            plt.plot(context_switches[0], context_switches[1])
            plt.savefig("context_switches.png")
            plt.show()

            # Compute statistics on context switches
            print("samples: %d" % len(context_switches[1]))
            print("mean: %f" % np.mean(context_switches[1]))
            print("stdev: %f" % np.std(context_switches[1]))
            print("variance: %f" % np.var(context_switches[1]))

            return None

        self.add_phase("baseline", baseline)


def main():
    s = PerfTimelineExperiment()
    s.run()


if __name__ == "__main__":
    main()
