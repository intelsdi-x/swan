class Taskset:
    def __init__(self, cpus, command):
        self.command = command
        self.cpus = cpus

    def __str__(self):
        return "taskset -c " + (",".join(self.cpus)) + " " + str(self.command)
