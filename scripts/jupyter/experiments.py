from cassandra.cluster import Cluster
import test_data_reader

class Experiments:
    def __init__(self, cassandra_cluster=None, test_file=None):
        self.cluster = None
        self.session = None
        self.experiments = {}

        rows = []
        if cassandra_cluster is not None:
            self.cluster = Cluster(cassandra_cluster)
            self.session = self.cluster.connect('snap')

            # Really inefficient queries do to the table layout :(
            rows = self.session.execute('SELECT tags, time FROM metrics')

        elif test_file is not None:
            rows = test_data_reader.read(test_file)

        for row in rows:
            if 'swan_experiment' in row.tags:
                self.experiments[row.tags['swan_experiment']] = row.time

    def _repr_html_(self):
        html_out = ""
        html_out += "<table>"
        html_out += "<tr><th>Experiment id</th><th>Date</th></tr>"
        for experiment, time in self.experiments.iteritems():
            html_out += "<tr><td>%s</td><td>%s</td></tr>" % (experiment, time)

        html_out += "</table>"

        return html_out

    def experiment(self, id):
        return Experiment(id, self.cluster, self.session)
