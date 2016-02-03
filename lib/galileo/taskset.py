class Taskset:
    """
    Convenience class for wrapping commands in taskset i.e. taskset -c 0,1,2 <command>
    """

    def __init__(self, cpus, command):
        """
        Generates prefix command for task set command

        :param cpus: List of cpu ids to run command on
        :param command: command to run
        """
        self.command = str(command)

        assert isinstance(cpus, list)
        self.cpus = cpus

    def __str__(self):
        """
        :return: Full command with taskset prefixed.
        """
        if len(self.cpus) == 0:
            return self.command

        return "taskset -c %s %s" % (",".join(self.cpus), self.command)
