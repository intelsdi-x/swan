import collections
import glog as log
import numpy as np
import os
import time
import uuid


class Experiment:
    def __init__(self):
        self.phases = collections.OrderedDict()

    def add_phase(self, name, phase, matrix=None):
        self.phases[name] = phase

    def run(self):
        run_id = str(uuid.uuid4())
        log.info("started experiment run '" + run_id + "'")

        start_experiment = time.time()

        # Create sandbox for experiment run
        os.mkdir('/'.join(['data', run_id]))

        # TODO: Write system info and time to root of experiment run directory.

        for name, phase in self.phases.iteritems():
            # Create sandbox for phase
            os.mkdir('/'.join(['data', run_id, name]))

            start_phase = time.time()

            results = []
            log.info("started phase '" + name + "'")
            for iteration in range(0, 3):
                root = os.getcwd()
                work_dir = '/'.join(['data', run_id, name, "run_" + str(iteration)])
                # Change directory to sandbox
                os.mkdir(work_dir)
                os.chdir(work_dir)

                start_iteration = time.time()

                log.info("started phase '" + name + "' iteration " + str(iteration))
                result = phase(None)

                if result is not None:
                    results.append(result)

                log.info("ended phase '" + name + "' iteration " + str(iteration) + " in " + str(
                    time.time() - start_iteration) + " seconds")

                # Change back to root directory
                os.chdir(root)

            metrics = {}
            log.info("ended phase '" + name + "' in " + str(time.time() - start_phase) + " seconds")

            for result in results:
                for metric_name, metric in result.iteritems():
                    if metric_name not in metrics:
                        metrics[metric_name] = []

                    metrics[metric_name].append(metric)

            for metric_name, metrics in metrics.iteritems():
                log.info('phase \'' + name + '\' result: ' + metric_name + '(mean of ' + str(
                    len(metrics)) + ' runs): ' + str(np.mean(metrics)))

        log.info("ended experiment run '" + run_id + "' in " + str(time.time() - start_experiment) + " seconds")
