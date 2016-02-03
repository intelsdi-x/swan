import glog as log
import os
import sys
import numpy as np
from uuid import UUID

sys.path.append('../../lib/galileo')
import perf_counters


def valid_uuid(name):
    try:
        val = UUID(name, version=4)
    except ValueError:
        return False
    return True


def main():
    if len(sys.argv) != 2:
        log.warning("Usage: %s <experiment output directory>" % sys.argv[0])
        sys.exit(1)

    root = sys.argv[1]
    experiments = os.listdir(root)
    for experiment in experiments:
        if valid_uuid(experiment):
            log.warning(
                "Run %s against experiment output directory. Not entire data directory i.e. 'data/<uuid>/' not 'data/'" %
                sys.argv[0])
            sys.exit(1)

        runs = os.listdir('/'.join([root, experiment]))

        latencies = []
        ipcs = []

        for run in runs:
            perf_file = '/'.join([root, experiment, run, "perf.txt"])
            try:
                timeline = perf_counters.Timeline(perf_file)
            except IOError:
                log.warning("Missing sample for '%s'" % perf_file)
                continue

            instructions_cycles = timeline.filter_by_columns(['instructions', 'cycles'])
            total_instructions = 0
            total_cycles = 0
            for line in instructions_cycles:
                total_instructions += line[0]
                total_cycles += line[1]

            ipc = float(total_instructions) / float(total_cycles)
            ipcs.append(ipc)

            mutilate_output = '/'.join([root, experiment, run, "output.txt"])
            with open(mutilate_output, 'r') as f:
                for line in f:
                    components = line.split()
                    if len(components) == 9 and components[0] == "read":
                        latency_p99 = float(components[8])
                        latencies.append(latency_p99)

        if len(latencies) > 0 and len(ipcs) > 0:
            print("statistics for experiment '%s':" % experiment)
            print("\t\tmean\t\tstdev\t\tcount\tvariance\tmin\t\tmax")
            print("latency (us):\t%f\t%f\t%d\t%f\t%f\t%f" % (
            np.mean(latencies), np.std(latencies), len(latencies), np.var(latencies), min(latencies), max(latencies)))
            print("IPC:\t\t%f\t%f\t%d\t%f\t%f\t%f" % (
            np.mean(ipcs), np.std(ipcs), len(ipcs), np.var(ipcs), min(ipcs), max(ipcs)))
            print("")


if __name__ == "__main__":
    main()
