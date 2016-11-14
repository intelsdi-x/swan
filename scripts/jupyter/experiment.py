"""
This module contains the convience class to read experiment data and generate sensitivity
profiles. See profile.py for more information.
"""
import pandas as pd
import test_data_reader

from cassandra.cluster import Cluster
from cassandra.query import SimpleStatement


class Experiment(object):
    """
    The Experiment class works as a container for swan experiment metrics.
    The experiment id should either be found when running the experiment through the swan cli or
    from using the Experiments class.
    """
    def __init__(self, experiment_id, **kwargs):
        """
        :param experiment_id: string of experiment_id gathered from cassandra
        :param name: optional name of the experiment
        :param cassandra_cluster: ip of cassandra cluster in a string format
        :param port: endpoint od cassandra cluster [int]
        :param read_csv: if no specify cassandra_cluster and port, try to read from a csv

        Initializes an experiment from a given experiment id by using the cassandra cluster and
        session.
        Set cassandra_cluster to an array of hostnames where cassandra nodes reside.
        """
        self.rows = {}  # keep temporary rows from query for later match with qps rows
        self.qps = {}  # qps is a one row from query, where we can map it to percentile rows
        self.data = []
        self.experiment_id = experiment_id
        self.name = kwargs['name'] if 'name' in kwargs else self.experiment_id
        self.columns = ['ns', 'host', 'time', 'value', 'plugin_running_on', 'swan_loadpoint_qps', 'achieved_qps_percent',
                        'swan_experiment', 'swan_aggressor_name', 'swan_phase', 'swan_repetition']

        port = kwargs['port'] if 'port' in kwargs else 9042
        keyspace = kwargs['keyspace'] if 'keyspace' in kwargs else 'snap'
        if 'cassandra_cluster' in kwargs:
            self.cluster = Cluster(kwargs['cassandra_cluster'], port=port)
            self.session = self.cluster.connect(keyspace)
            self.match_qps()

        elif 'read_csv' in kwargs:
            self.rows, self.qps = test_data_reader.read(kwargs['read_csv'])

        self.populate_data()
        self.frame = pd.DataFrame(self.data, columns=self.columns)

    def match_qps(self):
        query = """SELECT ns, ver, host, time, boolval, doubleval, strval, tags, valtype
            FROM snap.metrics WHERE tags['swan_experiment'] = \'%s\'""" % self.experiment_id
        statement = SimpleStatement(query, fetch_size=100)
        for row_count, row in enumerate(self.session.execute(statement), start=1):
            k = (row.ns, row.tags['swan_aggressor_name'], row.tags['swan_phase'], row.tags['swan_repetition'])
            self.rows[k] = row
            if row.ns == "/intel/swan/%s/%s/qps" % (row.ns.split("/")[3], row.host):
                self.qps[(row.ns, row.tags['swan_phase'], row.tags['swan_repetition'])] = row.doubleval

    def populate_data(self):
        for row in self.rows.itervalues():
            if row.valtype == "doubleval":
                ns = "/intel/swan/%s/%s/qps" % (row.ns.split("/")[3], row.host)
                achived_qps = (self.qps[(ns, row.tags['swan_phase'], row.tags['swan_repetition'])] /
                               float(row.tags['swan_loadpoint_qps']))
                percent_qps = '{percent:.2%}'.format(percent=achived_qps)
                values = [row.ns, row.host, row.time, row.doubleval, row.tags['plugin_running_on'],
                          row.tags['swan_loadpoint_qps'], percent_qps, row.tags['swan_experiment'],
                          row.tags['swan_aggressor_name'], row.tags['swan_phase'], row.tags['swan_repetition']]

                self.data.append(values)

    def get_frame(self):
        return self.frame

    def _repr_html_(self):
        df_of_99th = self.frame.loc[self.frame['ns'].str.contains('/percentile/99th')]
        df_of_99th.groupby(['swan_aggressor_name', 'swan_phase', 'swan_repetition'])
        return df_of_99th.to_html(index=False)


if __name__ == '__main__':
    Experiment(experiment_id='57d25f69-d6d7-43e1-5c4e-3b5f5208acdc', cassandra_cluster=['127.0.0.1'], port=9042)
