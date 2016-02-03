import collections


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

        return "perf stat -x ',' --append %s -I %d -o %s %s" % (
            events_string, self.interval, self.output_file, self.command)


class TimelineEntry:
    def __init__(self):
        self.time = 0.0
        self.data = {}


class Timeline:
    def __init__(self, input_file):
        self.started = ""
        self.entries = []
        self.input_file = input_file
        self.process()

    def process(self):
        self.reset()

        # Parser assumes:
        # - No out of order entries.
        # - Start line with 'started at XYZ' appears only once

        with open(self.input_file, 'r') as f:
            found_start_tag = False
            next_entry = None

            for line in f:
                line = line.strip()
                if line == '':
                    continue

                if found_start_tag == False and '# started on ' in line:
                    self.started = line
                    found_start_tag = True
                    continue

                # Allow any inline comments.
                if line[0] == '#':
                    continue

                # Skip non supported counters
                cols = line.split(',')
                if cols[1] == '<not supported>':
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
        self.started = ""
        self.entries = []

    def filter_by_columns(self, columns, seperate_columns=False):
        column_lookup = {}
        for column in columns:
            column_lookup[column] = True

        output = []
        if seperate_columns == False:
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

    def output(self, separator='\t'):
        lines = []

        legend = []
        legend_lookup = {}

        for entry in self.entries:
            line = []
            line.append(str(entry.time))
            for metric, data_point in entry.data.iteritems():
                if metric not in legend_lookup:
                    # TODO: Verify metric column
                    legend_lookup[metric] = True
                    legend.append(metric)

                line.append(str(data_point))

            lines.append(separator.join(line))

        return ["#time" + separator + separator.join(legend)] + lines

    def tsv(self):
        return self.output('\t')

    def csv(self):
        return self.output(',')
