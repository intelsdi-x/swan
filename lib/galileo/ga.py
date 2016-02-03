import collections
import glog as log
import os
import time
import uuid


class Experiment:
    """
    Main driver for experiments. Use 'add_phase()' to add desired experiment phases and execute with 'run()'.
    """

    def __init__(self):
        self.run_id = None
        self.phases = collections.OrderedDict()

    def add_phase(self, name, phase, matrix=None):
        """
        Submits phase for experiment execution.

        :param name: Phase name
        :param phase: Function to execute during phase
        :param matrix: Test matrix. 2 dimensional array. For example [[A, B], [C, D]]
        """
        self.phases[name] = {'phase': phase, 'matrix': matrix}

    def generate_permutations(self, matrix):
        """
        Generates all permutations from test matrix.

        :param matrix: 2 dimensional array. For example [[A, B], [C, D]]
        :return: The above becomes: [[A, C], [A, D], [B, C], [B, D]]
        """
        permutations = []

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

        return permutations

    def run(self, repetitions=3):
        """
        Runs experiment

        :param repetitions: Number of times to repeat each phase.
        """
        if self.run_id is not None:
            log.warning("overriding run id '%s'" % self.run_id)

        self.run_id = str(uuid.uuid4())
        log.info("started experiment run '" + self.run_id + "'")

        start_experiment = time.time()

        # Ensure data dir is present
        if not os.path.exists('data/'):
            os.mkdir('data/')

        # Create sandbox for experiment run
        os.mkdir('/'.join(['data', self.run_id]))

        # TODO: Write system info and time to root of experiment run directory.

        for name, phase_tuple in self.phases.iteritems():
            # Generate permutations of phase.
            matrix = phase_tuple["matrix"]
            permutations = self.generate_permutations(matrix)
            phase = phase_tuple["phase"]

            for permutation in permutations:
                # Create sandbox for phase
                if permutation is None:
                    phase_dir = '/'.join(['data', self.run_id, name])
                else:
                    mangle_permutation = "_".join((map(str, permutation)))
                    phase_dir = '/'.join(['data', self.run_id, name + "_" + mangle_permutation])

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

                    log.info(
                        "started phase '" + name + "' configuration %s" % permutation + " iteration " + str(iteration))
                    phase(permutation)

                    log.info("ended phase '" + name + "' configuration %s" % permutation + " iteration " + str(
                        iteration) + " in " + str(
                        time.time() - start_iteration) + " seconds")

                    # Cool off
                    time.sleep(0.5)

                    # Change back to root directory
                    os.chdir(root)

                log.info("ended phase '" + name + "' configuration %s" % permutation + " in " + str(
                    time.time() - start_phase) + " seconds")

        log.info("ended experiment run '" + self.run_id + "' in " + str(time.time() - start_experiment) + " seconds")
