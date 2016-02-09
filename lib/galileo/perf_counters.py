import collections


class Perf:
    """
    Convenience class for recording perf statistics for a given command.
    """

    def __init__(self, command, events=None, interval=1000, output_file="perf.txt"):
        """
        :param command:     Command to wrap.
        :param events:      List of event names to collect. Is empty by default and plain `perf stat` will be run.
        :param interval:    Interval to collect (in milliseconds). Defaults to 1000 milliseconds.
                            If set to None, statistics will be written when command exits.
        :param output_file: File name to write statistics to. Defaults to "perf.txt"
        """
        self.command = str(command)
        self.interval = interval
        self.events = events
        self.output_file = output_file

    def __str__(self):
        events_string = ""
        if self.events is not None:
            events_string = ("-e %s " % ",".join(self.events))

        interval_string = ""
        if self.interval is not None:
            interval_string = ("-I %d " % self.interval)

        return "perf stat -x ',' --append %s%s-o %s %s" % (
            events_string, interval_string, self.output_file, self.command)


class TimelineEntry:
    """
    Helper container for Timeline records. I.e. a full sample of one or more data points for a given instant in time.
    These are traditionally split across multiple lines with `perf stat -I XXXX`.
    """

    def __init__(self):
        self.time = 0.0
        self.data = {}


class Timeline:
    """
    Class to help parse `perf stat` output files.
    It process the perf output and aggregates data points per time-step and enables subsequent filtering.
    """

    def __init__(self, input_file):
        """
        :param input_file: Perf output file to parse. For example, "perf.txt"
        """
        self.started = ""
        self.entries = []
        self.input_file = input_file
        self.process()

    def process(self):
        """
        Internal function to populate self.entries[] from self.input_file.
        """

        self.reset()

        # Parser assumes:
        # - No out of order entries e.g. samples will come in order: 0.012 X, 0.012 Y, 0.023 X, 0.023 Y, ...
        # - Start line with "started at XYZ" appears only once

        with open(self.input_file, "r") as f:
            found_start_tag = False
            next_entry = None

            for line in f:
                line = line.strip()
                if line == "":
                    continue

                if found_start_tag == False and "# started on " in line:
                    self.started = line
                    found_start_tag = True
                    continue

                # Allow any inline comments.
                if line[0] == "#":
                    continue

                # Skip non supported counters
                cols = line.split(",")
                if cols[1] == "<not supported>":
                    continue

                # Bootstrap with first entry
                if next_entry is None:
                    next_entry = TimelineEntry()
                    next_entry.time = float(cols[0])

                # Commit entry to entries if new timestamp is seen.
                if next_entry.time != float(cols[0]):
                    self.entries.append(next_entry)
                    next_entry = TimelineEntry()
                    next_entry.time = float(cols[0])

                # Add metrics to entry
                next_entry.data[cols[3]] = float(cols[1])

        # if next_entry isn't None: add it to entries
        if next_entry is not None:
            self.entries.append(next_entry)

    def reset(self):
        """
        Internal function to clean up previous parsings. Is called by process() on entry.
        """
        self.started = ""
        self.entries = []

    def filter_by_columns(self, columns, separate_columns=False):
        """
        Returns a subset and/or transformed columns/rows from self.entries.

        :param columns:             List of metrics to filter. Including 'time' in the list will include the time stamp
                                    in the row.
        :param separate_columns:    If set to False (default), output will be row based.
                                    For example:
                                        Entries [{time: 0.12, data: {A: 10, B: 20}}, {time: 0.23, data: {A: 15, B: 25}}]
                                        With 'columns' ["time", "B"]
                                        Becomes [[0.12, 20], [0.23, 25]]

                                    If set to True, output will be column based.
                                        Same example as above becomes [[0.12, 0.23], [20, 25]]
        :return:                    Two dimensional array either row or column based (determined by 'separate_columns')
                                    with data points filtered by 'columns'.
        """
        column_lookup = {}
        for column in columns:
            column_lookup[column] = True

        output = []
        if separate_columns == False:
            for entry in self.entries:
                output_entry = []
                if "time" in column_lookup:
                    output_entry.append(entry.time)

                for metric, data_point in entry.data.iteritems():
                    if metric in column_lookup:
                        output_entry.append(data_point)

                output.append(output_entry)
        else:
            columns = collections.OrderedDict()
            for entry in self.entries:
                if "time" in column_lookup:
                    if "time" not in columns:
                        columns["time"] = []

                    columns["time"].append(entry.time)

                for metric, data_point in entry.data.iteritems():
                    if metric in column_lookup:
                        if metric not in columns:
                            columns[metric] = []

                        columns[metric].append(data_point)

            for name, column in columns.iteritems():
                output.append(column)

        return output

    def output(self, separator):
        """
        Traverse entries and flattens into a list of strings separated by `separator`.
        :param separator: Column separator to use. For example '\t' for tab separated output.
        :return:          list of strings separated by `separator`, prefixed with comment line describing the columns.
                          For example: ["#time,A,B", "0.12,10,20", "0.23,15,25"]
        """
        lines = []

        legend = []
        legend_lookup = {}

        for entry in self.entries:
            line = [str(entry.time)]
            for metric, data_point in entry.data.iteritems():
                if metric not in legend_lookup:
                    # TODO: Verify metric column
                    legend_lookup[metric] = True
                    legend.append(metric)

                line.append(str(data_point))

            lines.append(separator.join(line))

        return ["#time" + separator + separator.join(legend)] + lines

    def tsv(self):
        """
        :return: Tab separated serialization of self.entries. See output() for more information on output format.
        """
        return self.output("\t")

    def csv(self):
        """
        :return: Comma separated serialization of self.entries. See output() for more information on output format.
        """
        return self.output(",")
