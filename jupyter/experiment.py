"""
This module contains the convience class to read experiment data and generate sensitivity
profiles. See profile.py for more information.
"""
import ssl
import os.path

from collections import defaultdict

import numpy as np
import pandas as pd

from cassandra.auth import PlainTextAuthProvider
from cassandra.cluster import Cluster
from cassandra.query import SimpleStatement

import test_data_reader


class Experiment(object):
    """
    The Experiment class works as a container for swan experiment metrics.
    The experiment id should either be found when running the experiment through the swan cli or
    from using the Experiments class.
    """

    CASSANDRA_SESSION = None  # one instance for all existing notebook experiments

    @staticmethod
    def _create_or_get_session(cassandra_cluster, port, ssl_options, keyspace):
        if not Experiment.CASSANDRA_SESSION:
            auth_provider = None
            if ssl_options:
                if 'ssl_version' not in ssl_options:
                    context = ssl.SSLContext(ssl.PROTOCOL_SSLv23)
                    context.options |= ssl.OP_NO_SSLv2
                    context.options |= ssl.OP_NO_SSLv3
                    ssl_options['ssl_version'] = context.protocol

                auth_provider = PlainTextAuthProvider(username=ssl_options['username'],
                                                      password=ssl_options['password'])

            cluster = Cluster(cassandra_cluster, port=port, ssl_options=ssl_options, auth_provider=auth_provider)
            Experiment.CASSANDRA_SESSION = cluster.connect(keyspace)

        return Experiment.CASSANDRA_SESSION

    def __init__(self, experiment_id, cassandra_cluster=None, port=9042, name=None, keyspace='snap',
                 aggressor_throughput_namespaces_prefixes=(), ssl_options=None, read_csv=None, cached=True):
        """
        :param experiment_id: string of experiment_id gathered from cassandra
        :param name: optional name of the experiment if missing, experiment_id is given instead
        :param cassandra_cluster: ip of cassandra cluster in a string format
        :param port: endpoint od cassandra cluster [int]
        :param keyspace: keyspace used in cassandra cluster
        :param aggressor_throughput_namespaces_prefixes: get work done for aggressor by specify ns prefix in cassandra
        :param ssl_options used during secure connection.
            Keys needed are: `ca_certs`, `keyfile`, `certfile` which are absolute paths,
                             `username` with `password` are mandatory,
                             `ssl_version` is optional and created in case of missing
        :param read_csv: if no specify cassandra_cluster and port, try to read from a csv
        :param cached: pickle experiment to data directory for offline usage

        Initializes an experiment from a given experiment id by using the cassandra cluster and
        session.
        Set cassandra_cluster to an array of hostnames where cassandra nodes reside.
        """
        self.experiment_id = experiment_id
        cached_experiment = os.path.join('data', '%s.pkl' % self.experiment_id)
        if cached and os.path.exists(cached_experiment):
            self.frame = pd.read_pickle(cached_experiment)
            return

        self.throughputs = defaultdict(list)  # keep throughputs from all aggressors to join it later with main DF
        self.aggressor_throughput_namespaces_prefixes = aggressor_throughput_namespaces_prefixes

        self.name = name if name else self.experiment_id
        self.columns = ['ns', 'host', 'time', 'value', 'plugin_running_on', 'swan_loadpoint_qps',
                        'achieved_qps_percent', 'swan_experiment', 'swan_aggressor_name', 'swan_phase',
                        'swan_repetition', 'throughputs']

        if cassandra_cluster:
            session = Experiment._create_or_get_session(cassandra_cluster, port, ssl_options, keyspace)
            rows, qps = self.match_qps(session)

        elif read_csv:
            rows, qps = test_data_reader.read(read_csv)

        data = self.populate_data(rows, qps)

        self.frame = pd.DataFrame(data, columns=self.columns)
        if cached and not os.path.exists(cached_experiment):
            os.makedirs('data') if not os.path.exists('data') else None
            self.frame.to_pickle(cached_experiment)

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
            elif filter(lambda ns: ns in row.ns, self.aggressor_throughput_namespaces_prefixes):
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

                max_throughput = max(self.throughputs[k]) if k in self.throughputs else np.nan

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
