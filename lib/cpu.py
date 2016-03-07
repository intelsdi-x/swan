import random


def unique(source, count):
    """
    :param source: Map of sockets, cores or hyper threads.
    :param count: Number of unique items to sample.
    :return: A list of unique non-overlapping items from source map.
    """
    keys = random.sample(source, count)
    output = []
    for key in keys:
        output.append(source[key])
    return output


class HyperThread:
    def __init__(self, id):
        """
        :param id: Processor Id for hyper threaded core.
                   If hyper threads are not enabled,
                   there will be a 1:1 mapping between hyper thread and core structure.
        """
        self.id = id
        self.raw_data = {}


class Core:
    """
    Represents a processor core. Cores consist of hyper threads.
    If hyper threads are disabled, the hyper thread count is 1.
    """

    def __init__(self, id):
        """
        :param id: Core Id
        """
        self.id = id
        self.hyper_threads = {}

    def unique_hyper_threads(self, count):
        """
        :param count: Number of hyper threads.
        :return: A list of "count" unique (non overlapping) hyper threads.
                 If count is greater than the available hyper threads, all hyper threads are returned.
        """
        return unique(self.hyper_threads, count)


class Socket:
    """
    Represents physical processor die (with it's own socket).
    Sockets contains a number of "cores", which then contains the actual "hyper threads".
    """

    def __init__(self, id):
        """
        :param id: Physical Processor Id
        """
        self.id = id
        self.cores = {}

    def unique_cores(self, count):
        """
        :param count: Number of cores.
        :return: A list of "count" unique (non overlapping) cores. If count is greater than the available cores,
                 all cores are returned.
        """
        return unique(self.cores, count)


class Cpus:
    """
    Represents all available sockets, cores and hyper threads available on the system.
    """

    def __init__(self, cpu_info_file="/proc/cpuinfo"):
        """
        :param cpu_info_file: cpuinfo file to use i.e. /proc/cpuinfo
               Usually, this should be left alone and is only made configurable for testing purposes.
        """
        self.hyper_threads = {}
        self.sockets = {}
        next_cpu = None

        with open(cpu_info_file) as f:
            for line in f:
                components = line.rstrip("\n").split(":")
                if len(components) != 2:
                    continue
                key = components[0].strip()
                value = components[1].strip()

                if key == "processor":
                    if next_cpu is not None:
                        self.hyper_threads[next_cpu.id] = next_cpu

                    next_cpu = HyperThread(int(value))

                next_cpu.raw_data[key] = value

            if next_cpu is not None:
                self.hyper_threads[next_cpu.id] = next_cpu

        for hyper_thread_id, hyper_thread in self.hyper_threads.items():
            if "physical id" not in hyper_thread.raw_data:
                continue

            socket_id = int(hyper_thread.raw_data["physical id"])
            if socket_id not in self.sockets:
                self.sockets[socket_id] = Socket(socket_id)

            core_id = int(hyper_thread.raw_data["core id"])
            if core_id not in self.sockets[socket_id].cores:
                self.sockets[socket_id].cores[core_id] = Core(core_id)

            if hyper_thread.id not in self.sockets[socket_id].cores[core_id].hyper_threads:
                self.sockets[socket_id].cores[core_id].hyper_threads[hyper_thread.id] = hyper_thread

    def unique_sockets(self, count):
        """
        :param count: Number of sockets.
        :return: A list of "count" unique (non overlapping) sockets. If count is greater than the available sockets,
                 all sockets are returned.
        """
        return unique(self.sockets, count)
