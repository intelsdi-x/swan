import ga
import glog as log


class SimpleExperiment(ga.Experiment):
    def __init__(self):
        ga.Experiment.__init__(self)

        def baseline(configuration):
            return None

        def experiment(configuration):
            log.info("Test configuration: " + str(configuration))
            return None

        self.add_phase("baseline", baseline)
        self.add_phase("experiment", experiment, [["cpu", "mem"], range(0, 8), range(100, 1000, 100)])


def main():
    s = SimpleExperiment()
    s.run()


if __name__ == "__main__":
    main()