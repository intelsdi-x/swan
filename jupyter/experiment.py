# encoding: utf-8

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
from cassandra.query import named_tuple_factory, ordered_dict_factory


pd.set_option('max_colwidth',400)


class Experiment(object):
    """
    The Experiment class works as a container for swan experiment metrics.
    The experiment id should either be found when running the experiment through the swan cli or
    from using the Experiments class.
    """

    CASSANDRA_SESSION = None  # one instance for all existing notebook experiments

    @staticmethod
    def _create_or_get_session(cassandra_cluster, port, ssl_options):
        if not Experiment.CASSANDRA_SESSION:
            auth_provider = None
            if ssl_options:
                username = ssl_options.pop('username')
                password = ssl_options.pop('password')
                auth_provider = PlainTextAuthProvider(username, password)

            cluster = Cluster(cassandra_cluster, port=port, ssl_options=ssl_options, auth_provider=auth_provider)
            Experiment.CASSANDRA_SESSION = cluster.connect()

        return Experiment.CASSANDRA_SESSION

    def __init__(self, experiment_id, cassandra_cluster=None, port=9042, name=None, keyspaces=('snap', 'swan'),
                 aggressor_throughput_namespaces_prefixes=(), ssl_options=None, read_csv=False, dir_csv='data'):
        """
        :param experiment_id: string of experiment_id gathered from cassandra
        :param name: optional name of the experiment if missing, experiment_id is given instead
        :param cassandra_cluster: ip of cassandra cluster in a string format
        :param port: endpoint od cassandra cluster [int]
        :param keyspaces: keyspaces used in cassandra cluster. First is for data, second is for metadata.
        :param aggressor_throughput_namespaces_prefixes: get work done for aggressor by specify ns prefix in cassandra
        :param ssl_options used during secure connection.
            Keys needed are: `ca_certs`, `keyfile`, `certfile` which are absolute paths,
                             `username` with `password` are mandatory,
                             `ssl_version` is optional and created in case of missing
        :param read_csv: try to read  an experiment from from a csv in `dir_csv`
        :param dir_csv: save csv experiment to data directory for offline usage

        Initializes an experiment from a given experiment id by using the cassandra cluster and
        session.
        Set cassandra_cluster to an array of hostnames where cassandra nodes reside.
        """
        if not os.path.exists(dir_csv):
            os.makedirs(dir_csv)

        self.name = name if name else experiment_id
        self.cached_experiment = os.path.join(dir_csv, '%s.csv' % experiment_id)
        self.cached_meta_data_experiment = os.path.join(dir_csv, '%s_meta.csv' % experiment_id)

        if read_csv:
            self.frame = pd.read_csv(self.cached_experiment)
            self._meta = pd.read_csv(self.cached_meta_data_experiment)
        else:
            data, meta = self.read_data_from_cassandra(cassandra_cluster, port, keyspaces, ssl_options, experiment_id)

            self._meta = pd.DataFrame.from_dict(meta[0], orient='index')

            rows, qps, throughputs = self.match_qps(data, aggressor_throughput_namespaces_prefixes)
            self.frame = self.create_df(rows, qps, throughputs)

    def match_qps(self, response, aggressor_throughput_namespaces_prefixes):
        rows = {}  # keep temporary rows from query for later match with qps rows
        qps = {}  # qps is a one row from query, where we can map it to percentile rows
        throughputs = defaultdict(list)  # keep throughputs from all aggressors to join it later with main DF

        for row in response:
            k = (row.tags['swan_aggressor_name'], row.tags['swan_phase'], row.tags['swan_repetition'])
            if row.ns == "/intel/swan/%s/%s/qps" % (row.ns.split("/")[3], row.host):
                qps[k] = row.doubleval
            elif filter(lambda ns: ns in row.ns, aggressor_throughput_namespaces_prefixes):
                throughputs[k].append(row.doubleval)
            else:
                rows[(row.ns,) + k] = row
        return rows, qps, throughputs

    def read_data_from_cassandra(self, cassandra_cluster, port, keyspaces, ssl_options, experiment_id):
        session = Experiment._create_or_get_session(cassandra_cluster, port, ssl_options)

        query_for_data = """SELECT ns, ver, host, time, boolval, doubleval, strval, tags, valtype
                            FROM %s.metrics
                            WHERE tags['swan_experiment'] = \'%s\'  ALLOW FILTERING""" % (keyspaces[0], experiment_id)
        query_for_meta = """SELECT metadata
                            FROM %s.metadata
                            WHERE experiment_id = \'%s\' ALLOW FILTERING""" % (keyspaces[1], experiment_id)

        response_data = self.__execute_query(query_for_data, session, named_tuple_factory)
        response_meta = self.__execute_query(query_for_meta, session, ordered_dict_factory)

        return response_data, response_meta

    def __execute_query(self, query, session, row_factory):
        session.row_factory = row_factory
        statement = SimpleStatement(query, fetch_size=100)
        response = session.execute(statement)
        return response

    def create_df(self, rows, qps, throughputs):
        data = []
        for row in rows.itervalues():
            if row.valtype == "doubleval":
                k = (row.tags['swan_aggressor_name'], row.tags['swan_phase'], row.tags['swan_repetition'])

                expectedQPS = float(row.tags['swan_loadpoint_qps'])
                if expectedQPS == 0.0:
                    percent_qps = '+∞'
                else:
                    achieved_qps = qps[k] / expectedQPS
                    percent_qps = '{percent:.2%}'.format(percent=achieved_qps)

                max_throughput = max(throughputs[k]) if k in throughputs else np.nan

                values = [row.ns, row.host, row.time, row.doubleval, row.tags['plugin_running_on'],
                          row.tags['swan_loadpoint_qps'], percent_qps, row.tags['swan_experiment'],
                          row.tags['swan_aggressor_name'], row.tags['swan_phase'], row.tags['swan_repetition'],
                          max_throughput]

                data.append(values)

        columns = ['ns', 'host', 'time', 'value', 'plugin_running_on', 'swan_loadpoint_qps',
                   'achieved_qps_percent', 'swan_experiment', 'swan_aggressor_name', 'swan_phase',
                   'swan_repetition', 'throughputs']
        frame = pd.DataFrame(data, columns=columns)

        if not os.path.exists(self.cached_experiment):
            frame.to_csv(self.cached_experiment)
            self._meta.to_csv(self.cached_meta_data_experiment)

        return frame

    def get_frame(self):
        return self.frame

    def get_metadata(self):
        return self._meta

    def _repr_html_(self):
        df_of_99th = self.frame.loc[self.frame['ns'].str.contains('/percentile/99th')]
        df_of_99th.groupby(['swan_aggressor_name', 'swan_phase', 'swan_repetition'])
        return df_of_99th.to_html(index=False)


if __name__ == '__main__':
    Experiment(experiment_id='7be3c448-4fa2-4178-75aa-e23d292d4030', cassandra_cluster=['127.0.0.1'],
               port=9042, read_csv=False)
