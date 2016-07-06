from IPython.core.display import display, HTML
from cassandra.cluster import Cluster
import json
from sets import Set

class Sample:
    def __init__(self, ns, ver, host, time, boolval, doubleval, strval, tags, valtype):
        self.ns = ns
        self.ver = ver
        self.host = host
        self.time = time
        self.boolval = boolval
        self.doubleval = doubleval
        self.strval = strval
        self.tags = tags
        self.valtype = valtype

    def metric_name(self):
        # Sanitize metric name by stripping namespace prefix.
        # For example, from '/intel/swan/mutilate/jp7-1/percentile/99th' to
        # 'percentile/99th'.

        # Make sure it is a metric we recognise.
        if not self.ns.startswith('/intel/swan/mutilate/'):
            return self.ns

        namespace_exploded = self.ns.split('/')

        # Make sure we have at least '/intel/swan/mutilate/', the host name and a metric.
        if len(namespace_exploded) < 5:
            return self.ns

        metric_exploded = namespace_exploded[5:]
        return '/'.join(metric_exploded)



class Experiment:
    def __init__(self, id, cluster, session):
        self.id = id
        self.cluster = cluster
        self.session = session

        self.samples = []
        self.phases = {}
        self.sensivity_rows = {}
        self.phase_to_aggressor = {}

        lookup = self.session.prepare('SELECT ns, ver, host, time, boolval, doubleval, strval, tags, valtype FROM snap.metrics WHERE tags CONTAINS ? ALLOW FILTERING')
        sample_rows = self.session.execute(lookup, [self.id])
        for sample_row in sample_rows:
            sample = Sample(sample_row.ns, sample_row.ver, sample_row.host, sample_row.time, sample_row.boolval, sample_row.doubleval, sample_row.strval, sample_row.tags, sample_row. valtype)

            # Categorize in phase and sample for sorting and lookup.
            if 'swan_phase' in sample_row.tags:
                if sample_row.tags['swan_phase'] not in self.phases:
                    self.phases[sample_row.tags['swan_phase']] = []
                self.phases[sample_row.tags['swan_phase']].append(sample)

                # Categorize for sensitivity profile.
                if 'swan_loadpoint_qps' in sample_row.tags:
                    phase = sample_row.tags['swan_phase']

                    # Ugly hack. Phase names are repeated and specialized from Swan side.
                    # Should be fixed when we introduce configuration unfolding.
                    phase = phase.split('_id_')[0]

                    if 'swan_aggressor_name' in sample_row.tags:
                        self.phase_to_aggressor[phase] = sample_row.tags['swan_aggressor_name']

                    load_point = int(sample_row.tags['swan_loadpoint_qps'])

                    if phase not in self.sensivity_rows:
                        self.sensivity_rows[phase] = {}

                    if load_point not in self.sensivity_rows[phase]:
                        self.sensivity_rows[phase][load_point] = {}

                    self.sensivity_rows[phase][load_point][sample.metric_name()] = sample

            self.samples.append(sample)

    def show_samples(self):
        html_out = ""
        html_out += "<table>"
        html_out += "<tr><th>Phase</th><th>Repetition</th><th>Metric</th><th>Value</th></tr>"
        for phase, repetitions in self.phases.iteritems():
            # Times two is a mega hack. Should be removed.
            phase_column = "<td rowspan=%d>%s</td>" % (len(repetitions) * 2, phase)
            for sample in repetitions:
                repetition = 0
                if 'swan_repetition' in sample.tags:
                    repetition = sample.tags['swan_repetition']

                html_out += "<tr>%s<td>%s</td><td>%s</td><td>%s</td><tr>" % (phase_column, repetition, sample.ns, sample.doubleval)
                phase_column = ""

        html_out += "</table>"

        display(HTML(html_out))

    # SLO should be read from database.
    def show_sensitivity_profile(self, SLO):
        # HTML styling constants
        no_border = "border: 0"
        black_border = "1px solid black; "

        html_out = ''
        html_out += '<table style="border: 0;">'
        html_out += '<tr style="%s">' % no_border
        html_out += '<th style="%s; border-bottom: %s; border-right: %s;">Scenario / Load</th>' % (no_border, black_border, black_border)

        for load_percentage in range(5, 100, 10):
            html_out += '<th style="border: 0; border-bottom: 1px solid black;">%s%%</th>' % load_percentage

        html_out += '</tr>'

        for phase in self.sensivity_rows:
            html_out += '<tr style="%s">' % no_border

            aggressor = self.phase_to_aggressor[phase]
            if aggressor == "None":
                label = "Baseline"
            else:
                label = aggressor

            html_out += '<td style="%s; border-right: %s;">%s</td>' % (no_border, black_border, label)

            # Yet another hack. We have to sort the load points from lowest to highest.
            sorted_loadpoints = []
            for load_points in self.sensivity_rows[phase]:
                sorted_loadpoints.append(load_points)
            sorted_loadpoints.sort()

            for load_points in sorted_loadpoints:
                samples = self.sensivity_rows[phase][load_points]

                if 'percentile/99th' in samples:
                    latency = samples['percentile/99th']
                    violation = ((latency.doubleval / SLO) * 100)
                    style = "%s; " % no_border

                    if violation > 150:
                        style += "background-color: #a9341f; color: white;"
                    elif violation > 100:
                        style += "background-color: #ffeda0;"
                    else:
                        style += "background-color: #98cc70;"

                    html_out += '<td style="%s">%.1f%%</td>' % (style, violation)
        else:
            html_out += '<td style="%s"></td>' % no_border

            html_out += '</tr>'

        html_out += '</table>'

        display(HTML(html_out))

class Experiments:
    def __init__(self, cassandra_cluster):
        self.cluster = None
        self.session = None

        self.cluster = Cluster(cassandra_cluster)
        self.session = self.cluster.connect('snap')

    def list(self):
        # Really inefficient queries do to the table layout :(
        experiments = {}
        rows = self.session.execute('SELECT tags, time FROM metrics')
        for row in rows:
            if 'swan_experiment' in row.tags:
                experiments[row.tags['swan_experiment']] = row.time

        return experiments

    def show_list(self):
        experiments = self.list()

        html_out = ""
        html_out += "<table>"
        html_out += "<tr><th>Experiment id</th><th>Date</th></tr>"
        for experiment, time in experiments.iteritems():
            html_out += "<tr><td>%s</td><td>%s</td></tr>" % (experiment, time)

        html_out += "</table>"

        display(HTML(html_out))

    def experiment(self, id):
        return Experiment(id, self.cluster, self.session)
