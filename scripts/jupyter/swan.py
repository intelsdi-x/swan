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

                    self.sensivity_rows[phase][load_point][sample_row.ns] = sample

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
        html_out = ''
        html_out += '<table>'
        html_out += '<tr>'
        html_out += '<th>Scenario / Load</th>'

        for load_percentage in range(5, 100, 10):
            html_out += '<th>%s%%</th>' % load_percentage

        html_out += '</tr>'

        for phase in self.sensivity_rows:
            html_out += '<tr>'

            aggressor = self.phase_to_aggressor[phase]
            if aggressor == "None":
                label = "Baseline"
            else:
                label = aggressor

            html_out += '<td>%s</td>' % label

            # Yet another hack. We have to sort the load points from lowest to highest.
            sorted_loadpoints = []
            for load_points in self.sensivity_rows[phase]:
                sorted_loadpoints.append(load_points)
            sorted_loadpoints.sort()

            for load_points in sorted_loadpoints:
                samples = self.sensivity_rows[phase][load_points]
                latency = samples["/intel/swan/mutilate/localhost.localdomain/percentile/99th"]
                violation = ((latency.doubleval / SLO) * 100)

                if violation > 150:
                    color = "red"
                elif violation > 100:
                    color = "orange"
                else:
                    color = "green"

                html_out += '<td style="background-color: %s;">%.1f%%</td>' % (color, violation)

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
