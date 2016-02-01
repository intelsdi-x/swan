class HyperThread:
    def __init__(self, id):
        self.id = id
        self.raw_data = {}

class Core:
    def __init__(self, id):
        self.id = id
        self.hyper_threads = {}

class Socket:
    def __init__(self, id):
        self.id = id
        self.cores = {}

class Cpus:
    def __init__(self):
        self.hyper_threads = {}
        self.sockets = {}
        next_cpu = None

        with open('/proc/cpuinfo') as f:
            for line in f:
                components = line.rstrip('\n').split(':')
                if len(components) != 2:
                    continue
                key = components[0].strip()
                value = components[1].strip()

                if key == 'processor':
                    if next_cpu != None:
                        self.hyper_threads[next_cpu.id] = next_cpu

                    next_cpu = HyperThread(int(value))
                
                next_cpu.raw_data[key] = value 

            if next_cpu != None:
                self.hyper_threads[next_cpu.id] = next_cpu

        for hyper_thread_id, hyper_thread in self.hyper_threads.iteritems():
            if 'physical id' not in hyper_thread.raw_data:
                continue

            socket_id = int(hyper_thread.raw_data['physical id'])
            if socket_id not in self.sockets:
                self.sockets[socket_id] = Socket(socket_id)

            core_id = int(hyper_thread.raw_data['core id'])
            if core_id not in self.sockets[socket_id].cores:
                self.sockets[socket_id].cores[core_id] = Core(core_id)

            if hyper_thread.id not in self.sockets[socket_id].cores[core_id].hyper_threads:
                self.sockets[socket_id].cores[core_id].hyper_threads[hyper_thread.id] = hyper_thread

        print '%d sockets' % len(self.sockets)
        for socket_id, socket in self.sockets.iteritems():
            print 'Socket %d has %d cores' % (socket.id, len(socket.cores))
            for core_id, core in socket.cores.iteritems():
                print '\tCore %d has %d hyper threads' % (core.id, len(core.hyper_threads))
                for hyper_thread_id, hyper_thread in core.hyper_threads.iteritems():
                    print '\t\tHyper thread id: %d' % hyper_thread.id
