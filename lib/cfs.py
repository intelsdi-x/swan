import glog as log

class Policy:
    """
    Represents CFS scheduling policies.
    """
    def __init__(self, flag, realtime=False):
        self.flag = flag
        self.realtime = realtime

    def __str__(self):
        return self.flag


class CFS:
    """
    Convenience class to control the scheduling policy and scheduling
    priority of a command.
    """

    SCHED_OTHER = Policy('--other')
    SCHED_IDLE = Policy('--idle')
    SCHED_BATCH = Policy('--batch')
    SCHED_RR = Policy('--rr', True)
    SCHED_FIFO = Policy('--fifo', True)

    def __init__(self, policy, priority, command):
        """
        :param policy:   CFS scheduling policy to use to run the command.
                         Valid policy values are "BATCH", "FIFO", "IDLE",
                         "OTHER", and "RR". OTHER is the default CFS policy,
                         so it doesn't make much sense to use that option.
        :param priority: The scheduling priority. Defaults to 1 if real-time,
                         and 0 otherwise. Non-zero values are valid only for
                         the real-time policies (RR, FIFO). On many Linux
                         systems the maximum priority is 99. You can determine
                         the min and max values on your system by running
                         `chrt --max`.
        :param command:  Command to run.
        """
        if priority is None:
            priority = 0
        self.priority = priority
        self.policy = policy
        self.cmd = command
        if policy.realtime and priority < 1:
            raise Exception('Priority for realtime policy must be positive')
        if not policy.realtime and priority != 0:
            raise Exception('Priority for normal policy must be 0')

    def __str__(self):
        return "chrt %s %d %s" % (self.policy, self.priority, self.cmd)
