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
        self.phases[name] = {'phase': phase, 'matrix': matrix}

    def run(self, repetitions=3):
        run_id = str(uuid.uuid4())
        log.info("started experiment run '" + run_id + "'")

        start_experiment = time.time()

        # Create sandbox for experiment run
        os.mkdir('/'.join(['data', run_id]))

        # TODO: Write system info and time to root of experiment run directory.

        for name, phase_tuple in self.phases.iteritems():
            # Generate permutations of phase.
            permutations = []
            matrix = phase_tuple["matrix"]

            # O(n^2): watch out for large matrices
            if matrix is not None:
                def perm(m, permutation_fragment):
                    if len(m) <= 0:
                        permutations.append(permutation_fragment)
                    else:
                        for item in m[0]:
                            perm(m[1:], permutation_fragment + [item])
                perm(matrix, [])
            else:
                permutations = [None]

            phase = phase_tuple["phase"]

            for permutation in permutations:
                # Create sandbox for phase
                phase_dir = ""
                if permutation is None:
                    phase_dir = '/'.join(['data', run_id, name])
                else:
                    mangle_permutation = "_".join((map(str, permutation)))
                    phase_dir = '/'.join(['data', run_id, name + mangle_permutation])

                os.mkdir(phase_dir)

                start_phase = time.time()

                results = []
                log.info("started phase '" + name + "' configuration %s" % permutation)
                for iteration in range(0, repetitions):
                    root = os.getcwd()

                    work_dir = '/'.join([phase_dir, "run_" + str(iteration)])
                    # Change directory to sandbox
                    os.mkdir(work_dir)
                    os.chdir(work_dir)

                    start_iteration = time.time()

                    log.info("started phase '" + name + "' configuration %s" % permutation + " iteration " + str(iteration))
                    result = phase(permutation)

                    if result is not None:
                        results.append(result)

                    log.info("ended phase '" + name + "' configuration %s" % permutation + " iteration " + str(iteration) + " in " + str(
                        time.time() - start_iteration) + " seconds")

                    # Change back to root directory
                    os.chdir(root)

                metrics = {}
                log.info("ended phase '" + name + "' configuration %s" % permutation + " in " + str(time.time() - start_phase) + " seconds")

                for result in results:
                    for metric_name, metric in result.iteritems():
                        if metric_name not in metrics:
                            metrics[metric_name] = []

                        metrics[metric_name].append(metric)

                for metric_name, metrics in metrics.iteritems():
                    log.info('phase \'' + name + '\' result: ' + metric_name + '(mean of ' + str(
                        len(metrics)) + ' runs): ' + str(np.mean(metrics)))

        log.info("ended experiment run '" + run_id + "' in " + str(time.time() - start_experiment) + " seconds")
