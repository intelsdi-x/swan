"""
This module contains the convience class to read experiment data and generate sensitivity
profiles. See profile.py for more information.
"""
import pandas as pd
import test_data_reader

from cassandra.cluster import Cluster
from cassandra.query import SimpleStatement
from collections import defaultdict


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
        :param aggressor_throughput_namespaces_prefix: get work done for aggressor by specify ns prefix in cassandra DB

        Initializes an experiment from a given experiment id by using the cassandra cluster and
        session.
        Set cassandra_cluster to an array of hostnames where cassandra nodes reside.
        """
        self.throughputs = defaultdict(list)  # keep throughputs from all aggressors to join it later with main DF
        self.experiment_id = experiment_id

        self.name = kwargs['name'] if 'name' in kwargs else self.experiment_id

        self.aggressor_throughput_namespaces_prefix = ()
        if 'aggressor_throughput_namespaces_prefix' in kwargs:
            self.aggressor_throughput_namespaces_prefix = kwargs['aggressor_throughput_namespaces_prefix']

        self.columns = ['ns', 'host', 'time', 'value', 'plugin_running_on', 'swan_loadpoint_qps',
                        'achieved_qps_percent', 'swan_experiment', 'swan_aggressor_name', 'swan_phase',
                        'swan_repetition', 'throughputs']

        port = kwargs['port'] if 'port' in kwargs else 9042
        keyspace = kwargs['keyspace'] if 'keyspace' in kwargs else 'snap'
        if 'cassandra_cluster' in kwargs:
            cluster = Cluster(kwargs['cassandra_cluster'], port=port)
            session = cluster.connect(keyspace)
            rows, qps = self.match_qps(session)

        elif 'read_csv' in kwargs:
            rows, qps = test_data_reader.read(kwargs['read_csv'])

        data = self.populate_data(rows, qps)
        self.frame = pd.DataFrame(data, columns=self.columns)

    def match_qps(self, session):
        rows = {}  # keep temporary rows from query for later match with qps rows
        qps = {}  # qps is a one row from query, where we can map it to percentile rows
        query = """SELECT ns, ver, host, time, boolval, doubleval, strval, tags, valtype
            FROM snap.metrics WHERE tags['swan_experiment'] = \'%s\'""" % self.experiment_id
        statement = SimpleStatement(query, fetch_size=100)

        for row in session.execute(statement):
            k = (row.tags['swan_aggressor_name'], row.tags['swan_phase'], row.tags['swan_repetition'])
            if row.ns == "/intel/swan/%s/%s/qps" % (row.ns.split("/")[3], row.host):
                qps[k] = row.doubleval
            elif filter(lambda ns: ns in row.ns, self.aggressor_throughput_namespaces_prefix):
                self.throughputs[k].append(row.doubleval)
            else:
                rows[(row.ns,) + k] = row
        return rows, qps

    def populate_data(self, rows, qps):
        data = []
        for row in rows.itervalues():
            if row.valtype == "doubleval":
                k = (row.tags['swan_aggressor_name'], row.tags['swan_phase'], row.tags['swan_repetition'])

                achieved_qps = (qps[k] / float(row.tags['swan_loadpoint_qps']))
                percent_qps = '{percent:.2%}'.format(percent=achieved_qps)

                max_throughput = max(self.throughputs[k]) if k in self.throughputs else None

                values = [row.ns, row.host, row.time, row.doubleval, row.tags['plugin_running_on'],
                          row.tags['swan_loadpoint_qps'], percent_qps, row.tags['swan_experiment'],
                          row.tags['swan_aggressor_name'], row.tags['swan_phase'], row.tags['swan_repetition'],
                          max_throughput]

                data.append(values)
        return data

    def get_frame(self):
        return self.frame

    def _repr_html_(self):
        df_of_99th = self.frame.loc[self.frame['ns'].str.contains('/percentile/99th')]
        df_of_99th.groupby(['swan_aggressor_name', 'swan_phase', 'swan_repetition'])
        return df_of_99th.to_html(index=False)


if __name__ == '__main__':
    Experiment(experiment_id='57d25f69-d6d7-43e1-5c4e-3b5f5208acdc', cassandra_cluster=['127.0.0.1'], port=9042)
