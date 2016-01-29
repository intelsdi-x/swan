import ga
from perf_counters import Perf
from shell import Shell
from taskset import Taskset
from cgroup import Cgroup


class CgroupExperiment(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        def baseline(configuration):
            cg = Cgroup(["/A/cpu.shares=8192", "/A/hp/cpu.shares=1024", "/A/be/cpu.shares=2"])

            Shell([
                cg.execute("/A/hp", Perf(Taskset(["0"], "echo foobar"))),
                cg.execute("/A/be", Perf(Taskset(["0"], "echo foobar")))
            ])

            cg.destroy()

            return None

        def experiment(configuration):
            Shell([Perf(Taskset(["0"], "sleep 1"))])

            return None

        self.add_phase("baseline", baseline)
        self.add_phase("experiment", experiment)


def main():
    s = CgroupExperiment()
    s.run()


if __name__ == "__main__":
    main()
