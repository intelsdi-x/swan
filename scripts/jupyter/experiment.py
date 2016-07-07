from cassandra.cluster import Cluster
import test_data_reader
from sample import Sample
from profile import Profile

class Experiment:
    def __init__(self, id, cluster=None, session=None, test_file=None):
        self.id = id
        self.cluster = cluster
        self.session = session

        self.samples = []
        self.phases = {}

        sample_rows = []
        if cluster is not None and session is not None:
            lookup = self.session.prepare('SELECT ns, ver, host, time, boolval, doubleval, strval, tags, valtype FROM snap.metrics WHERE tags CONTAINS ? ALLOW FILTERING')
            sample_rows = self.session.execute(lookup, [self.id])
        elif test_file is not None:
            sample_rows = test_data_reader.read(test_file)

        for sample_row in sample_rows:
            sample = Sample(sample_row.ns, sample_row.ver, sample_row.host, sample_row.time, sample_row.boolval, sample_row.doubleval, sample_row.strval, sample_row.tags, sample_row. valtype)

            # Categorize in phase and sample for sorting and lookup.
            if 'swan_phase' in sample_row.tags:
                if sample_row.tags['swan_phase'] not in self.phases:
                    self.phases[sample_row.tags['swan_phase']] = []
                self.phases[sample_row.tags['swan_phase']].append(sample)

            self.samples.append(sample)

    def profile(self, SLO):
        return Profile(self.samples, SLO)

    def _repr_html_(self):
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

        return html_out
