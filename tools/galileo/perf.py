class Perf:
    def __init__(self, command, events=None, interval=1000, output_file='perf.txt'):
        self.command = str(command)
        self.interval = interval
        self.events = events
        self.output_file = output_file

    def __str__(self):
        events_string = ""
        if self.events is not None:
            events_string = (" -e %s" % ",".join(self.events))

        return "perf stat --append %s -I %d -o %s %s" % (events_string, self.interval, self.output_file, self.command)
