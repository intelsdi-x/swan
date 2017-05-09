"""
 Copyright (c) 2017 Intel Corporation

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.


This module contains the convince class to read experiment data and generate sensitivity
profiles. See profile.py for more information.
"""

import sys
import re
import os
import datetime
import pickle
from functools import partial
import pandas as pd
import numpy as np


# -------------------------------
# dataframe on disk caching layer
# -------------------------------
class DataFrameToCSVCache:
    """ CSV based cache to store dataframe as compressed CSV."""

    CACHE_DIR = '.experiments_csv_cache'

    def _filename(self, experiment_id):
        return os.path.join(self.CACHE_DIR, '%s.cache' % experiment_id)

    def __contains__(self, experiment_id):
        return os.path.exists(self._filename(experiment_id))

    def __setitem__(self, experiment_id, df):
        if not os.path.exists(self.CACHE_DIR):
            os.makedirs(self.CACHE_DIR)
        df.to_csv(self._filename(experiment_id), compression='bz2')

    def __getitem__(self, experiment_id):
        if experiment_id in self:
            return pd.read_csv(self._filename(experiment_id), compression='bz2')
        else:
            raise KeyError()

    def __call__(self, fun):
        """ Can be use as decorator for function that have experiment_id as first parameter and returns dataframe."""
        def _dec(experiment_id, *args, **kw):
            if experiment_id in self:
                return self[experiment_id]

            df = fun(experiment_id, *args, **kw)

            self[experiment_id] = df
            return df
        return _dec


# -------------------------------
# cassandra session singleton
# -------------------------------
CASSANDRA_SESSION = None  # one instance for all existing notebook experiments
DEFAULT_CASSANDRA_OPTIONS = dict(
    nodes=['127.0.0.1'],
    port=9042,
    ssl_options=None
)


def _get_or_create_cassandra_session(nodes, port, ssl_options=None):
    """ Get or prepare new session to Cassandra cluster.

    :param nodes: List of addresses of cassandra nodes.
    :param port: Port of cassandra service listening on.
    :param ssl_options: Optional SSL options to connect to cassandra in secure manner.
    :returns: cassandra session singleton

    """
    from cassandra.auth import PlainTextAuthProvider
    from cassandra.cluster import Cluster
    from cassandra.query import ordered_dict_factory
    global CASSANDRA_SESSION
    if not CASSANDRA_SESSION:
        auth_provider = None
        if ssl_options:
            username = ssl_options.pop('username')
            password = ssl_options.pop('password')
            auth_provider = PlainTextAuthProvider(username, password)

        cluster = Cluster(nodes, port=port, ssl_options=ssl_options, auth_provider=auth_provider)
        CASSANDRA_SESSION = cluster.connect()
        CASSANDRA_SESSION.row_factory = ordered_dict_factory
    return CASSANDRA_SESSION


# ----------------------------------------
# helper function to load and convert data
# ----------------------------------------


def _load_rows_from_cassandra(experiment_id, cassandra_session, keyspace='snap'):
    """ Load data from cassandra database as rows.

    It will cache loaded rows from cassandra to speed up further requests.
    It only loads doubleval, ns and tags values ignoring other kinds of types.

    :returns: Data as cassandra rows (each row being dict like).
    """

    from cassandra.query import SimpleStatement

    query = """SELECT ns, doubleval, tags
               FROM %s.metrics
               WHERE tags['swan_experiment'] = \'%s\'  ALLOW FILTERING""" % (
                   keyspace, experiment_id)

    statement = SimpleStatement(query)

    print ("loading data from database...")
    started = datetime.datetime.now()

    rows = list(cassandra_session.execute(statement))
    if len(rows) == 0:
        print >>sys.stderr, "no metrics found!"
        return []
    print ("loaded %d rows in %0.fs" % (len(rows), (datetime.datetime.now() - started).seconds))

    return rows


def strip_prefix(column):
    """ Drop "/intel/swan/PLUGIN/HOST/" prefix from 'ns' column.

    E.g.
    /intel/swan/caffe/inference/host-123.asdfdf/batches -> batches
    /intel/swan/mutilate/ovh3-2342.123/percentile/99th -> percentile/99th
    /intel/swan/mutilate/ovh3-2342.123/std -> std

    Thanks to this user is not required to specify host name to gather metrics using unique experiment id.
    """
    pattern = re.compile(r'(/intel/swan/(caffe/)?(\w+)/([.\w-]+)/).*?')
    return column.apply(
        partial(pattern.sub, '')
    )


def _convert_rows_to_dataframe(rows):
    """ Convert rows (dicts) representing snap.metrics to pandas.DataFrame.

    E.g. transform raw row (dict like) data:

    rows = [ {
                 'ns':'/intel/swan/$PLUGIN/%HOST/percentile/99th',
                 'doubleval': 124,
                 'tags':{'tag1':'v1', 'tag2':'v2'}
             },,,    ]


    into dataframe:

    | ns                                        | doubleval | tag1 | tag2 |
    -----------------------------------------------------------------------
    | /intel/swan/$PLUGIN/$HOST/percentile/99th | 124       | v1   | v2   |

    :returns: dataframe and common tag keys which were found on every metric (intersection)
    """

    tag_keys = None

    # flat & map tags into new columns and find common tags in all rows
    for row in rows:
        tags = row.pop('tags')

        if tag_keys is None:
            tag_keys = set(tags.keys())  # Initially fill with first found tags.
        else:
            tag_keys = tag_keys.intersection(tags.keys())

        row.update(tags)

    return pd.DataFrame.from_records(rows), tag_keys


def _reshape_ns_to_columns(df, tag_keys, aggfunc=np.mean):
    """ Transform ns + doubleval columns into "ns" separate columns.

    :param df: Input raw dataframe with "ns" column and doubleval.
    :param tag_keys: Names of columns used to aggregate by.
    :param aggfunc: Function to aggregate the same values for the same tag_keys.

    For example, with input dataframe like this:

    | ns                                        | doubleval | tag1 | tag2 |
    -----------------------------------------------------------------------
    | /intel/swan/$PLUGIN/$HOST/percentile/99th | 124       | v1   | v2   |
    | /intel/swan/$PLUGIN/$HOST/std             | 515       | v1   | v2   |
    | /intel/swan/$PLUGIN/$HOST/qps             | 20000     | v1   | v2   |


    returns:

    | percentile/99th | std | qps   | tag1 | tag2 |
    -----------------------------------------------
    | 124             | 515 | 20000 | v1   | v2   |
    """

    # Drop prefix from 'ns' column (make independent from 'host' component).
    df['ns'] = strip_prefix(df['ns'])

    # Convert all series to numeric if possible.
    for column in df.columns:
        try:
            df[column] = df[column].apply(pd.to_numeric)
        except ValueError:
            continue

    # Reshape - to have have all tags + "ns" column as index.
    # First step: tags and ns column is converted to index with group by.
    groupby_columns = list(tag_keys)+['ns']
    grouper = df.groupby(groupby_columns)
    # and then just take a value from 'doubleval' column.
    value_grouper = grouper['doubleval']
    # Get a mean value of 'doubleval' column values grouped by tags and ns.
    # df = value_grouper.mean()  # TODO: consider making this an option (cpu vs batches issue).
    df = value_grouper.aggregate(aggfunc)
    # Then use 'ns' categorical column values will become new columns.
    df = df.unstack(['ns'])
    # Reset index to transform "tag_keys" columns used by group by to become ordinary columns.
    df = df.reset_index()

    return df


@DataFrameToCSVCache()
def load_dataframe_from_cassandra(experiment_id, cassandra_options, aggfunc=np.mean):
    """ Basic generic function to load experiment data from cassandra with rows grouped by given tags.

    Return dateframe is cached by DataFrameToCSVCache

    :param experiment_id: identifier of experiment to load data for,
    :param cassandra_options: identifier of experiment to load data for,
    :param tag_keys: tag keys used to group doubleval values,
    :param aggfunc: what function used to aggregate values with the same tags,
    :returns: pandas.Dataframe with all tags and ns categorical values as columns (check convert_rows_to_dataframe
              for details).

    """
    cassandra_session = _get_or_create_cassandra_session(**(cassandra_options or DEFAULT_CASSANDRA_OPTIONS))

    rows = _load_rows_from_cassandra(experiment_id, cassandra_session)
    df, tag_keys = _convert_rows_to_dataframe(rows)

    return _reshape_ns_to_columns(df, tag_keys, aggfunc=aggfunc)


# Constants for columns names used.
# New column names.


ACHIEVED_QPS_LABEL = 'achieved QPS'
ACHIEVED_LATENCY_LABEL = 'achieved latency'
COMPOSITE_VALUES_LABEL = 'composite values'

# Existing column names (from metrics provided by plugins).
PERCENTILE99TH_LABEL = 'percentile/99th'
NUMBER_OF_CORES = 'number_of_cores'
QPS_LABEL = 'qps'  # Absolute achieved QPS.
SWAN_LOAD_POINT_QPS_LABEL = 'swan_loadpoint_qps'  # Target QPS.
SWAN_AGGRESSOR_NAME_LABEL = 'swan_aggressor_name'
SNAP_USE_COMPUTER_SATURATION_LABEL = '/intel/use/compute/saturation'

# --------------------------------
# Style functions for single cells
# --------------------------------


CRIT = 'background:#a9341f; color: white;'
WARN = 'background:#ffeda0'
OK = 'background:#98cc70'
NAN = 'background-color: #c0c0c0'


def composite_qps_formatter(composite_values, normalized=False, show_qps=False):
    if composite_values is None:
        return 'N/A'

    qps = composite_values[QPS_LABEL]
    achieved_qps = composite_values[ACHIEVED_QPS_LABEL]

    if any(map(pd.isnull, (achieved_qps, qps))):
        return NAN

    if normalized:
        return '{:.0%}'.format(achieved_qps)
    else:
        return '{:,.0f}'.format(qps)


def composite_qps_colors(composite_values, show_qps=False):
    if pd.isnull(composite_values):
        return NAN

    achieved_qps = composite_values[ACHIEVED_QPS_LABEL]
    if any(map(pd.isnull, (achieved_qps, ))):
        return NAN

    # format just according QPS value
    if pd.isnull(achieved_qps):
        return NAN
    if achieved_qps > 0.95:
        return OK
    elif achieved_qps > 0.8:
        return WARN
    else:
        return CRIT


def composite_latency_formatter(composite_values, normalized=False):
    if composite_values is None:
        return 'N/A'

    achieved_qps = composite_values[ACHIEVED_QPS_LABEL]
    latency = composite_values[PERCENTILE99TH_LABEL]
    achieved_latency = composite_values[ACHIEVED_LATENCY_LABEL]

    if any(map(pd.isnull, (achieved_qps, latency, achieved_latency))):
        return NAN

    if achieved_qps < 0.9:
        return 'FAIL'

    if normalized:
        if achieved_latency > 3.0:
            return '> 300%'
        else:
            return '{:.0%}'.format(achieved_latency)
    else:
        return '{:.0f}'.format(latency)


def composite_latency_colors(composite_values, slo):
    if pd.isnull(composite_values):
        return NAN

    achieved_qps = composite_values[ACHIEVED_QPS_LABEL]
    latency = composite_values[PERCENTILE99TH_LABEL]

    if any(map(pd.isnull, (achieved_qps, latency))):
        return NAN

    # format just according QPS value
    if pd.isnull(achieved_qps) or pd.isnull(latency):
        return NAN

    if achieved_qps < 0.9:
        return NAN

    if latency > 1.5 * slo:
        return CRIT
    elif latency > slo:
        return WARN
    else:
        return OK


class Experiment:
    """ Base class for experiments that are based on load(QPS) and latency.

    Responsible for loading data from cassandra, reshaping and preparing extra normalized columns
    for latency/qps according target SLOs.

    Extra normalized columns are then stored in one composite column for further processing.

    Additionally provides a helper function to rename existing dataframe columns to new names and
    function that allows to refere to new names using original names (_renamed).
    """

    def __init__(self, experiment_id, slo, cassandra_options=None, columns_to_rename=None, aggfunc=np.mean):
        df = load_dataframe_from_cassandra(experiment_id, cassandra_options, aggfunc=aggfunc)
        df = self.add_extra_and_composite_columns(df, slo)
        self.columns_to_rename = columns_to_rename
        self.experiment_id = experiment_id
        self.slo = slo
        self.df = df

    def _rename_columns(self, df):
        """ Rename columns according self.columns_to_rename.
            Use _renamed method to refer to new names. """
        return self.df.rename(columns=self.columns_to_rename)

    def _renamed(self, original_name):
        """ Returns new name of column to be used by formatting & styling functions. """
        return self.columns_to_rename.get(original_name, original_name)

    def add_extra_and_composite_columns(self, df, slo):
        """ Add extra derived columns with achieved normalized QPS/latency
        and composite column to store all values in one dict."""
        # Extra columns.
        # Calculate achieved QPS as percentage (normalized to 1).
        df[ACHIEVED_QPS_LABEL] = pd.Series(df[QPS_LABEL] / df[SWAN_LOAD_POINT_QPS_LABEL])

        # Calculate achieved latency in regards to SLO.
        df[ACHIEVED_LATENCY_LABEL] = pd.Series(df[PERCENTILE99TH_LABEL] / slo)

        # Columns to store in one cell for father processing.
        COMPOSITE_COLUMNS = [ACHIEVED_QPS_LABEL, PERCENTILE99TH_LABEL, ACHIEVED_LATENCY_LABEL, QPS_LABEL]
        # Composite value to store all values e.g. "achieved qps" and "latency" together in one cell as dict.
        # Used to display one of the values and format using other value.
        df[COMPOSITE_VALUES_LABEL] = df[COMPOSITE_COLUMNS].apply(dict, axis=1)

        return df


class OptimalCoreAllocation(Experiment):
    """ Experiment that presents latency/QPS and cpu utilization in "number of cores" and
        "load" dimensions.
    """
    def __init__(self, experiment_id, slo, cassandra_options=None):
        Experiment.__init__(self, experiment_id, slo, cassandra_options, columns_to_rename={
            NUMBER_OF_CORES: 'Number of cores',
            SWAN_LOAD_POINT_QPS_LABEL: 'Target QPS',
        },
            aggfunc=np.max,  # for cpu utilization,
        )

    def _composite_pivot_table(self):
        df = self._rename_columns(self.df)
        return df.pivot_table(
                values=COMPOSITE_VALUES_LABEL,
                index=self._renamed(NUMBER_OF_CORES),
                columns=self._renamed(SWAN_LOAD_POINT_QPS_LABEL),
                aggfunc='first',
            )

    def _get_caption(self, cell, normalized):
        return "Optimal core allocation: %s%s of %s experiment" % (
            'normalized ' if normalized else '',
            cell,
            self.experiment_id
        )

    def latency(self, normalized=True):
        return self._composite_pivot_table(
            ).style.applymap(
                partial(composite_latency_colors, slo=self.slo),
            ).format(
                partial(composite_latency_formatter, normalized=normalized)
            ).set_caption(
                self._get_caption('latency[us]', normalized)
            )

    def qps(self, normalized=True):
        return self._composite_pivot_table(
            ).style.applymap(
                partial(composite_qps_colors),
            ).format(
                partial(composite_qps_formatter, normalized=normalized)
            ).set_caption(
                self._get_caption('queries per second', normalized)
            )

    def cpu(self):
        def cpu_colors(cpu):
            """ Style function for cpu colors. """
            if pd.isnull(cpu):
                return NAN
            return "background: rgb(%d, %d, 0); color: white;" % (cpu * 255, 255 - cpu * 255)

        df = self._rename_columns(self.df)
        return df.pivot_table(
                values=SNAP_USE_COMPUTER_SATURATION_LABEL,
                index=self._renamed(NUMBER_OF_CORES),
                columns=self._renamed(SWAN_LOAD_POINT_QPS_LABEL),
                aggfunc=np.max,
            ).style.applymap(
                cpu_colors
            ).format(
                '{:.0%}'
            ).set_caption(
                self._get_caption('cpu utilization', False)
            )


class SensitivityProfile(Experiment):
    """ Experiment that presents latency/QPS and caffe aggressor throughput  in "aggressor" and
        "load" dimensions.
    """

    def __init__(self, experiment_id, slo, cassandra_options=None):
        Experiment.__init__(self, experiment_id, slo, cassandra_options, columns_to_rename={
            SWAN_LOAD_POINT_QPS_LABEL: 'Target QPS',
            SWAN_AGGRESSOR_NAME_LABEL: 'Aggressor',
        },
            aggfunc=np.max,  # for cpu utilization,
        )

        # Replace "None" aggressor with "Baseline" only for aggressor based experiments.
        self.df[SWAN_AGGRESSOR_NAME_LABEL].replace(to_replace={'None': 'Baseline'}, inplace=True)

    def _composite_pivot_table(self):
        df = self._rename_columns(self.df)
        return df.pivot_table(
                values=COMPOSITE_VALUES_LABEL,
                index=self._renamed(SWAN_AGGRESSOR_NAME_LABEL),
                columns=self._renamed(SWAN_LOAD_POINT_QPS_LABEL),
                aggfunc='first',
            )

    def _get_caption(self, cell, normalized=False):
        return "Sensitivity table: %s%s of %s experiment" % (
            'normalized ' if normalized else '',
            cell,
            self.experiment_id
        )

    def latency(self, normalized=True):
        return self._composite_pivot_table(
            ).style.applymap(
                partial(composite_latency_colors, slo=self.slo),
            ).format(
                partial(composite_latency_formatter, normalized=normalized)
            ).set_caption(
                self._get_caption('latency[us]', normalized)
            )

    def qps(self, normalized=True):
        return self._composite_pivot_table(
            ).style.applymap(
                partial(composite_qps_colors),
            ).format(
                partial(composite_qps_formatter, normalized=normalized)
            ).set_caption(
                self._get_caption('queries per second', normalized)
            )

    def caffe_batches(self):
        df = self._rename_columns(self.df)

        # For caffe batches only show caffe aggressor row.
        df = df[df[self._renamed(SWAN_AGGRESSOR_NAME_LABEL)] == 'Caffe']

        return df.pivot_table(
                values="batches",
                index=self._renamed(SWAN_AGGRESSOR_NAME_LABEL),
                columns=self._renamed(SWAN_LOAD_POINT_QPS_LABEL),
            ).style.format(
                '{:.0f}'
            ).set_caption(
                self._get_caption('caffe image batches')
            )
