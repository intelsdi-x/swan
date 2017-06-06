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
from collections import defaultdict


# -------------------------------
# dataframe on disk caching layer
# -------------------------------
class DataFrameToCSVCache:
    """ Dict-like API CSV based cache to store dataframe as compressed CSV.
    Can be used as decorator.
    """

    CACHE_DIR = '.experiments_csv_cache'

    def __init__(self, suffix=''):
        self.suffix = suffix

    def _filename(self, experiment_id):
        return os.path.join(self.CACHE_DIR, '%s%s.csv.bz2' % (experiment_id, self.suffix))

    def __contains__(self, experiment_id):
        return os.path.exists(self._filename(experiment_id))

    def __setitem__(self, experiment_id, df):
        if not os.path.exists(self.CACHE_DIR):
            os.makedirs(self.CACHE_DIR)
        df.to_csv(self._filename(experiment_id), compression='bz2', index=False)

    def __getitem__(self, experiment_id):
        if experiment_id in self:
            return pd.read_csv(self._filename(experiment_id), compression='bz2')
        else:
            raise KeyError()

    def __call__(self, func):
        """ Can be use as decorator for function that have experiment_id as first parameter and returns dataframe.
        Cache can be disabled by adding cache=False to kwargs in decorated function.
        """
        def decorator(experiment_id, *args, **kw):
            if kw.pop('cache', True) and experiment_id in self:
                return self[experiment_id]
            df = func(experiment_id, *args, **kw)
            self[experiment_id] = df
            return df
        return decorator


# -------------------------------
# cassandra session singleton
# -------------------------------
CASSANDRA_SESSION = None  # one instance for all existing notebook experiments
DEFAULT_CASSANDRA_OPTIONS = dict(
    nodes=['127.0.0.1'],
    port=9042,
    ssl_options=None
)
DEFAULT_KEYSPACE = 'swan'


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


@DataFrameToCSVCache()
def load_dataframe_from_cassandra_streamed(experiment_id, tag_keys, cassandra_options=DEFAULT_CASSANDRA_OPTIONS,
                                           aggfuncs=None, default_aggfunc=np.average,
                                           keyspace=DEFAULT_KEYSPACE):
    """ Load data from cassandra database as rows, processes them and returns dataframe with multiindex build on tags.

    It only loads doubleval, ns and tags values ignoring other kinds of metrics.

    :param experiment_id: identifier of experiment to load metrics for,
    :param tag_keys: Names of columns used to aggregate by and build Dataframe multiindex from.
    :param ns2agg: Mapping from processes "ns" to aggregation function for one phase.
    :param cassandra_session: Cassandra session object to be used to execute query,
    :param keyspace: Snap keyspace to be used to load metrics from,

    :returns: Data as cassandra rows (each row being dict like).
    """
    cassandra_session = _get_or_create_cassandra_session(**cassandra_options)

    # helper to drop prefix from ns (removing host depedency).
    pattern = re.compile(r'(/intel/swan/(caffe/)?(\w+)/([.\w-]+)/).*?')
    drop_prefix = partial(pattern.sub, '')

    query = "SELECT ns, doubleval, tags FROM %s.metrics WHERE tags['swan_experiment']=? ALLOW FILTERING" % keyspace
    statement = cassandra_session.prepare(query)

    rows = cassandra_session.execute(statement, [experiment_id])

    # temporary mutli hierarchy index for storing loaded data
    # first level is a namespace and second level is tuple of values from selected tags
    # value is is a list of values
    records = defaultdict(lambda: defaultdict(list))
    started = datetime.datetime.now()
    print('loading data...')
    idx = 0
    for idx, row in enumerate(rows):
        tags = row['tags']
        # namespace, value and tags
        ns = drop_prefix(row['ns'])
        val = row['doubleval']
        tagidx = tuple(tags[tk] for tk in tag_keys)
        # store in temporary index
        records[ns][tagidx].append(val)
        if idx and idx % 50000 == 0:
            print("%d loaded" % idx)
    print('%d rows loaded in %.0fs' % (idx, (datetime.datetime.now() - started).seconds))

    started = datetime.datetime.now()
    print('building a dataframe...')
    df = pd.DataFrame()
    for ns, d in records.items():
        tuples = []  # values used to build an index for given series.
        data = []  # data for Series
        # Use aggfunc provided by ns_aggfunctions or fallback to defaggfunc.
        aggfunc = aggfuncs.get(ns, default_aggfunc) if aggfuncs else default_aggfunc
        for tags, values in sorted(d.items()):
            tuples.append(tags)
            data.append(aggfunc(values))
        index = pd.MultiIndex.from_tuples(tuples, names=tag_keys)
        df[ns] = pd.Series(data, index)

    # Cannot recreate index after file is stored without additional info about number of index columns.
    # Additionally furhter transformation are based on values available in columns (not in index).
    df.reset_index(inplace=True)

    # Convert all series to numeric if possible.
    for column in df.columns:
        try:
            df[column] = df[column].apply(pd.to_numeric)
        except ValueError:
            continue

    print('dataframe with shape=%s build in %.0fs' % (df.shape, (datetime.datetime.now() - started).seconds))
    return df

# Constants for columns names used.
# New column names.


ACHIEVED_QPS_LABEL = 'achieved QPS'
ACHIEVED_LATENCY_LABEL = 'achieved latency'
COMPOSITE_VALUES_LABEL = 'composite values'

# Existing column names (from metrics provided by plugins).
PERCENTILE99TH_LABEL = 'percentile/99th'

QPS_LABEL = 'qps'  # Absolute achieved QPS.
SWAN_LOAD_POINT_QPS_LABEL = 'swan_loadpoint_qps'  # Target QPS.
SWAN_AGGRESSOR_NAME_LABEL = 'swan_aggressor_name'
SWAN_REPETITION_LABEL = 'swan_repetition'


# ----------------------------------------------------
# Style functions & constants for table cells styling
# ----------------------------------------------------


CRIT_STYLE = 'background:#a9341f; color: white;'
WARN_STYLE = 'background:#ffeda0'
OK_STYLE = 'background:#98cc70'
NAN_STYLE = 'background-color: #c0c0c0'
FAIL_STYLE = 'background-color: #b0c0b0'


QPS_FAIL_THRESHOLD = 0.9
QPS_OK_THRESHOLD = 0.95
QPS_WARN_THRESHOLD = 0.8

LATENCY_CRIT_THRESHOLD = 3  # 300 %
LATENCY_WARN_THRESHOLD = 1  # 100 %


def composite_qps_formatter(composite_values, normalized=False):
    """ Formatter responsible for showing either absolute or normalized value of QPS. """
    if composite_values is None:
        return 'N/A'

    qps = composite_values[QPS_LABEL]
    achieved_qps = composite_values[ACHIEVED_QPS_LABEL]

    if any(map(pd.isnull, (achieved_qps, qps))):
        return NAN_STYLE

    if normalized:
        return '{:.0%}'.format(achieved_qps)
    else:
        return '{:,.0f}'.format(qps)


def composite_qps_colors(composite_values):
    """ Styler for showing QPS from composite values.
    If not value available return NAN style.
    For achieved QPS > 95% show OK, with WARN color and CRITICAL color with
    achieved QPS above and below 80% respectively.
    """
    if pd.isnull(composite_values):
        return NAN_STYLE

    achieved_qps = composite_values[ACHIEVED_QPS_LABEL]
    if any(map(pd.isnull, (achieved_qps, ))):
        return NAN_STYLE

    if pd.isnull(achieved_qps):
        return NAN_STYLE
    if achieved_qps > QPS_OK_THRESHOLD:
        return OK_STYLE
    elif achieved_qps > QPS_WARN_THRESHOLD:
        return WARN_STYLE
    else:
        return CRIT_STYLE


def bytes_formatter(b):
    """ Formatter that formats bytes into kb/mb/gb etc... """
    for u in ' KMGTPEZ':
        if abs(b) < 1024.0:
            return "%3.1f%s" % (b, u)
        b /= 1024.0
    return "%.1f%s" % (b, 'Y')


def composite_latency_formatter(composite_values, normalized=False):
    """ Formatter responsible for showing either absolute or normalized value of latency depending of normalized argument.
    Additionally if achieved normalized QPS was below 90% marks column as "FAIL".
    """

    if composite_values is None:
        return 'N/A'

    achieved_qps = composite_values[ACHIEVED_QPS_LABEL]
    latency = composite_values[PERCENTILE99TH_LABEL]
    achieved_latency = composite_values[ACHIEVED_LATENCY_LABEL]

    if any(map(pd.isnull, (achieved_qps, latency, achieved_latency))):
        return NAN_STYLE

    if achieved_qps < QPS_FAIL_THRESHOLD:
        return '<b>FAIL</b>'

    if normalized:
        if achieved_latency > LATENCY_CRIT_THRESHOLD:
            return '>300%'
        else:
            return '{:.0%}'.format(achieved_latency)
    else:
        return '{:.0f}'.format(latency)


def composite_latency_colors(composite_values, slo):
    """ Styler responsible for choosing a background of latency tables.
    Uses composite value to get info about QPS and latency and then:
    - if normalized achieved QPS are below 90% return "fail style"
    - or depending of latency: if above 150% - CRIT, if above 100% WARN or OK otherwise
    """
    if pd.isnull(composite_values):
        return NAN_STYLE

    achieved_qps = composite_values[ACHIEVED_QPS_LABEL]
    achieved_latency = composite_values[ACHIEVED_LATENCY_LABEL]

    if any(map(pd.isnull, (achieved_qps, achieved_latency))):
        return NAN_STYLE

    # format just according QPS value
    if pd.isnull(achieved_qps) or pd.isnull(achieved_latency):
        return NAN_STYLE

    if achieved_qps < QPS_FAIL_THRESHOLD:
        return FAIL_STYLE

    if achieved_latency > LATENCY_CRIT_THRESHOLD:
        return CRIT_STYLE
    elif achieved_latency > LATENCY_WARN_THRESHOLD:
        return WARN_STYLE
    else:
        return OK_STYLE


class Renamer:

    """ Helper class to facilitate columns renaming in dataframe before visualizing.

    Instance can be used as function to refere to new name.
    """

    def __init__(self, columns_to_rename):
        self.columns_to_rename = columns_to_rename

    def rename(self, df):
        """ Rename columns according self.columns_to_rename.
            Use Renamer instance() method to refer to new names.

        :returns: new dataframe with columns renamed

        """
        return df.rename(columns=self.columns_to_rename)

    def __call__(self, original_name):
        """ Returns new name of column to be used by formatting & styling functions. """
        return self.columns_to_rename.get(original_name, original_name)


def add_extra_and_composite_columns(df, slo):
    """ Add extra derived columns with achieved normalized QPS/latency
    and composite column to store all values in one dict.

    Reshaping and preparing extra normalized columns
    for latency/qps according target SLOs.

    :returns: New dataframe with new derived columns and one special composite column.
    """
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


def _pivot_ui(df, totals=True, **options):
    """ Interactive pivot table for data analysis. """
    try:
        from pivottablejs import pivot_ui
    except ImportError:
        print("Error: cannot import pivottablejs, please install 'pip install pivottablejs'!")
        return
    iframe = pivot_ui(df, **options)
    if not totals:
        with open(iframe.src) as f:
            replacedHtml = f.read().replace(
                '</style>',
                '.pvtTotal, .pvtTotalLabel, .pvtGrandTotal {display: none}</style>'
            )
        with open(iframe.src, "w") as f:
            f.write(replacedHtml)
    return iframe


class Experiment:
    """ Base class for loading & storing data for swan experiments.

    During loading from cassandra data metrics are index and grouped by tag_keys (a.k.a. dimensions).
    Additionall parameter ns_aggfunctions is used to define how specific namespaces (ns) should be initially aggregated
    (by default it is 'mean' for whole phase).
    """

    def __init__(self, experiment_id, tag_keys, cassandra_options=DEFAULT_CASSANDRA_OPTIONS,
                 aggfuncs=None, default_aggfunc=np.mean, cache=True, keyspace=DEFAULT_KEYSPACE):
        self.experiment_id = experiment_id
        self.df = load_dataframe_from_cassandra_streamed(
            experiment_id, tag_keys, cassandra_options,
            aggfuncs=aggfuncs, default_aggfunc=default_aggfunc, cache=cache, keyspace=keyspace,
        )
        self.df.columns.name = 'Experiment %s' % self.experiment_id

    def _repr_html_(self):
        """ When presented in jupyter just return representation of dataframe. """
        return self.df._repr_html_()

    def pivot_ui(self):
        """ Interactive pivot table for data analysis. """
        return _pivot_ui(self.df)

# --------------------------------------------------------------
# "sensitivity profile" experiment
# --------------------------------------------------------------


class SensitivityProfile:
    """ Visualization for "sensitivity profile" experiments that presents
        latency/QPS and caffe aggressor throughput in "aggressor" and
        "load" dimensions.
    """

    tag_keys = (
        SWAN_AGGRESSOR_NAME_LABEL,
        SWAN_LOAD_POINT_QPS_LABEL,
        SWAN_REPETITION_LABEL,
    )

    def __init__(self, experiment_id, slo, cassandra_options=DEFAULT_CASSANDRA_OPTIONS,
                 cache=True, keyspace=DEFAULT_KEYSPACE):
        self.experiment = Experiment(experiment_id, self.tag_keys, cassandra_options,
                                     aggfuncs=dict(batches=np.max), cache=cache, keyspace=keyspace)
        self.slo = slo

        # Pre-process data specifically for this experiment.
        df = self.experiment.df.copy()
        df = add_extra_and_composite_columns(df, slo)

        # Replace "None" aggressor with "Baseline" only for aggressor based experiments.
        df[SWAN_AGGRESSOR_NAME_LABEL].replace(to_replace={'None': 'Baseline'}, inplace=True)

        # Rename columns
        self.renamer = Renamer({
            SWAN_LOAD_POINT_QPS_LABEL: 'Target QPS',
            SWAN_AGGRESSOR_NAME_LABEL: 'Aggressor',
        })
        self.df = self.renamer.rename(df)
        self.df.columns.name = 'Profile %s' % self.experiment.experiment_id

    def _repr_html_(self):
        """ When presented in jupyter just return representation of dataframe. """
        return self.df._repr_html_()

    def _composite_pivot_table(self, aggressors=None, qpses=None):
        df = self.df
        if aggressors is not None:
            df = df[df[self.renamer(SWAN_AGGRESSOR_NAME_LABEL)].isin(aggressors)]
        if qpses is not None:
            df = df[df[self.renamer(SWAN_LOAD_POINT_QPS_LABEL)].isin(qpses)]
        return df.pivot_table(
                values=COMPOSITE_VALUES_LABEL,
                index=self.renamer(SWAN_AGGRESSOR_NAME_LABEL),
                columns=self.renamer(SWAN_LOAD_POINT_QPS_LABEL),
                aggfunc='first',
            )

    def _get_caption(self, cell, normalized=False):
        return '%s%s of "sensitivity profile" experiment %s' % (
            'normalized ' if normalized else '',
            cell,
            self.experiment.experiment_id
        )

    def latency(self, normalized=True, aggressors=None, qpses=None):
        """ Generate table with information about tail latency."""
        return self._composite_pivot_table(
                aggressors,
                qpses
            ).style.applymap(
                partial(composite_latency_colors, slo=self.slo),
            ).format(
                partial(composite_latency_formatter, normalized=normalized)
            ).set_caption(
                self._get_caption('latency[us]', normalized)
            )

    def qps(self, normalized=True, aggressors=None, qpses=None):
        """ Generate table with information about achieved QPS."""
        return self._composite_pivot_table(
                aggressors,
                qpses
            ).style.applymap(
                partial(composite_qps_colors),
            ).format(
                partial(composite_qps_formatter, normalized=normalized)
            ).set_caption(
                self._get_caption('queries per second', normalized)
            )

    def caffe_batches(self):
        """ Generate table with information about Caffe aggressor
        images batches preprocessed darning each phase."""

        # For caffe only show caffe aggressor data.
        df = self.df[self.df[self.renamer(SWAN_AGGRESSOR_NAME_LABEL)] == 'Caffe']

        return df.pivot_table(
                values="batches",
                index=self.renamer(SWAN_AGGRESSOR_NAME_LABEL),
                columns=self.renamer(SWAN_LOAD_POINT_QPS_LABEL),
            ).style.format(
                '{:.0f}'
            ).set_caption(
                self._get_caption('caffe image batches')
            )

# --------------------------------------------------------------
# "optimal core allocation" experiment
# --------------------------------------------------------------


NUMBER_OF_CORES_LABEL = 'number_of_cores'  # HP cores. (TODO: replace with number_of_threads)
SNAP_USE_COMPUTE_SATURATION_LABEL = '/intel/use/compute/saturation'


class OptimalCoreAllocation:
    """ Visualization for "optimal core allocation" experiments that
        presents latency/QPS and cpu utilization in "number of cores" and "load" dimensions.
    """
    tag_keys = (
        SWAN_AGGRESSOR_NAME_LABEL,
        NUMBER_OF_CORES_LABEL,
        SWAN_LOAD_POINT_QPS_LABEL,
    )

    def __init__(self, experiment_id, slo, cassandra_options=DEFAULT_CASSANDRA_OPTIONS,
                 cache=True, keyspace=DEFAULT_KEYSPACE):
        self.experiment = Experiment(experiment_id, self.tag_keys, cassandra_options, cache=cache, keyspace=keyspace)
        self.slo = slo

        # Pre-process data specifically for this experiment.
        df = self.experiment.df.copy()
        df = add_extra_and_composite_columns(df, slo)

        # Rename columns.
        self.renamer = Renamer({
            SWAN_AGGRESSOR_NAME_LABEL: 'Aggressor',
            NUMBER_OF_CORES_LABEL: 'Number of cores',
            SWAN_LOAD_POINT_QPS_LABEL: 'Target QPS',
        })
        self.df = self.renamer.rename(df)
        self.df.columns.name = 'Optimal core allocation %s' % self.experiment.experiment_id

    def _repr_html_(self):
        return self.df._repr_html_()

    def _composite_pivot_table(self, aggressors=None, qpses=None):
        df = self.df
        if aggressors is not None:
            df = df[df[self.renamer(SWAN_AGGRESSOR_NAME_LABEL)].isin(aggressors)]
        if qpses is not None:
            df = df[df[self.renamer(SWAN_LOAD_POINT_QPS_LABEL)].isin(qpses)]
        return df.pivot_table(
                values=COMPOSITE_VALUES_LABEL,
                index=self.renamer(NUMBER_OF_CORES_LABEL),
                columns=self.renamer(SWAN_LOAD_POINT_QPS_LABEL),
                aggfunc='first',
            )

    def _get_caption(self, cell, normalized):
        return '%s%s of "optimal core allocation" experiment %s' % (
            'normalized ' if normalized else '',
            cell,
            self.experiment.experiment_id
        )

    def latency(self, normalized=True, aggressors=None, qpses=None):
        return self._composite_pivot_table(
                aggressors,
                qpses
            ).style.applymap(
                partial(composite_latency_colors, slo=self.slo),
            ).format(
                partial(composite_latency_formatter, normalized=normalized)
            ).set_caption(
                self._get_caption('latency[us]', normalized)
            )

    def qps(self, normalized=True, aggressors=None, qpses=None):
        return self._composite_pivot_table(
                aggressors,
                qpses
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
                return NAN_STYLE
            return "background: rgb(%d, %d, 0); color: white;" % (cpu * 255, 255 - cpu * 255)
        return self.df.pivot_table(
                values=SNAP_USE_COMPUTE_SATURATION_LABEL,
                index=self.renamer(NUMBER_OF_CORES_LABEL),
                columns=self.renamer(SWAN_LOAD_POINT_QPS_LABEL),
            ).style.applymap(
                cpu_colors
            ).format(
                '{:.0%}'
            ).set_caption(
                self._get_caption('cpu utilization', False)
            )


# --------------------------------------------------------------
# memcached-cat experiment
# --------------------------------------------------------------
def new_aggregated_index_based_column(df, source_indexes_column, template, aggfunc=sum):
    """ Create new pd.Series as aggregation of values from other columns.

    It uses template to find values from other columns, using indexes in one of the columns.
    E.g. with template='column-{}' and input dataframe like this:

    | example_indexes | column-1 | column-2 | column-3 |
    | 1,2             | 1        | 11       | 111      |
    | 1               | 2        | 22       | 222      |
    | 1,2,3           | 3        | 33       | 333      |

    when called like this
    >>> new_aggregate_cores_range_column('examples_indexes', template='column-{}', aggfunc=sum)

    results with series like this:

    | 12 (1+11)      |
    | 2              |
    | 369 (3+33+333) |

    """

    array = np.empty(len(df))
    for row_index, column_indexes in enumerate(df[source_indexes_column]):
        indexes = column_indexes.split(',')
        values = [df.iloc[row_index][template.format(index)] for index in indexes]
        aggvalue = aggfunc(values)
        array[row_index] = aggvalue
    return pd.Series(array)


# Derived metrics from Intel RDT collector.
LLC_BE_LABEL = 'llc/be/megabytes'
LLC_BE_PERC_LABEL = 'llc/be/perecentage'
MEMBW_BE_LABEL = 'membw/be/gigabytes'

LLC_HP_LABEL = 'llc/hp/megabytes'
LLC_HP_PERC_LABEL = 'llc/hp/perecentage'
MEMBW_HP_LABEL = 'membw/hp/gigabytes'

# BE configuration lables
BE_NUMBER_OF_CORES_LABEL = 'be_number_of_cores'
BE_L3_CACHE_WAYS_LABEL = 'be_l3_cache_ways'


class CAT:
    """ Visualization for "optimal core allocation" experiments that
        presents latency/QPS and cpu utilization in "number of cores" and "load" dimensions.
    """
    tag_keys = ('be_cores_range',
                'hp_cores_range',
                BE_L3_CACHE_WAYS_LABEL,
                BE_NUMBER_OF_CORES_LABEL,
                SWAN_AGGRESSOR_NAME_LABEL,
                SWAN_LOAD_POINT_QPS_LABEL)

    def __init__(self, experiment_id, slo, cassandra_options=DEFAULT_CASSANDRA_OPTIONS,
                 cache=True, keyspace=DEFAULT_KEYSPACE):

        self.experiment = Experiment(experiment_id, self.tag_keys, cassandra_options,
                                     aggfuncs=dict(batches=np.max), cache=cache, keyspace=keyspace)
        self.slo = slo

        df = self.experiment.df.copy()
        df = add_extra_and_composite_columns(df, slo)

        if '/intel/rdt/llc_occupancy/0/bytes' in df.columns:

            # aggregate BE columns
            df[LLC_BE_LABEL] = new_aggregated_index_based_column(
                df, 'be_cores_range', '/intel/rdt/llc_occupancy/{}/bytes', sum)/(1024*1024)
            df[MEMBW_BE_LABEL] = new_aggregated_index_based_column(
                df, 'be_cores_range', '/intel/rdt/memory_bandwidth/local/{}/bytes', sum)/(1024*1024*1024)

            df[LLC_HP_LABEL] = new_aggregated_index_based_column(
                df, 'hp_cores_range', '/intel/rdt/llc_occupancy/{}/bytes', sum)/(1024*1024)
            df[MEMBW_HP_LABEL] = new_aggregated_index_based_column(
                df, 'hp_cores_range', '/intel/rdt/memory_bandwidth/local/{}/bytes', sum)/(1024*1024*1024)

            df[LLC_BE_PERC_LABEL] = new_aggregated_index_based_column(
                df, 'be_cores_range', '/intel/rdt/llc_occupancy/{}/percentage', sum) / 100
            df[LLC_HP_PERC_LABEL] = new_aggregated_index_based_column(
                df, 'hp_cores_range', '/intel/rdt/llc_occupancy/{}/percentage', sum) / 100

        self.df = df

    def _get_caption(self, cell, normalized):
        return '%s%s of "memcached-cat" experiment %s' % (
            'normalized ' if normalized else '',
            cell,
            self.experiment.experiment_id
        )

    def filtered_df(self):
        """ Returns dataframe that exposes only meaningful columns."""

        # RDT collected data.
        rdt_columns = [
             LLC_HP_LABEL,
             LLC_HP_PERC_LABEL,

             LLC_BE_LABEL,
             LLC_BE_PERC_LABEL,

             MEMBW_HP_LABEL,
             MEMBW_BE_LABEL,
        ]

        columns = [
             SWAN_LOAD_POINT_QPS_LABEL,
             SWAN_AGGRESSOR_NAME_LABEL,
             BE_NUMBER_OF_CORES_LABEL,
             BE_L3_CACHE_WAYS_LABEL,
             PERCENTILE99TH_LABEL,
             ACHIEVED_LATENCY_LABEL,
             QPS_LABEL,
             ACHIEVED_QPS_LABEL,
        ]

        # Check if RDT collector data is available.
        if LLC_HP_LABEL in self.df:
            columns += rdt_columns

        df = self.df[columns]

        # Drop title of dataframe.
        df.columns.name = ''
        return df

    def filtered_df_table(self):
        """ Returns an simple formated dataframe """

        df = self.filtered_df()

        styler = df.style.format('{:.0%}', [
                ACHIEVED_QPS_LABEL,
                ACHIEVED_LATENCY_LABEL,
             ]
        )

        # format optionall values optionally.
        if LLC_HP_LABEL in self.df:
            styler = styler.format('{:.0%}', [
                LLC_BE_PERC_LABEL,
                LLC_HP_PERC_LABEL
             ])
        return styler

    def latency(self, normalized=True, aggressors=None, qpses=None):

        # Create local reference of data and modify it according provided paramters.
        df = self.df

        if aggressors is not None:
            df = df[df[SWAN_AGGRESSOR_NAME_LABEL].isin(aggressors)]

        if qpses is not None:
            df = df[df[SWAN_LOAD_POINT_QPS_LABEL].isin(qpses)]

        # Rename columns.
        renamer = Renamer({
            NUMBER_OF_CORES_LABEL: 'Number of cores',
            SWAN_LOAD_POINT_QPS_LABEL: 'Target QPS',
            BE_L3_CACHE_WAYS_LABEL: 'BE cache ways',
            BE_NUMBER_OF_CORES_LABEL: 'BE number of cores',
        })

        df = renamer.rename(df)

        return df.pivot_table(
                values=COMPOSITE_VALUES_LABEL, aggfunc='first',
                index=[renamer(SWAN_AGGRESSOR_NAME_LABEL), renamer(BE_L3_CACHE_WAYS_LABEL)],
                columns=[renamer(SWAN_LOAD_POINT_QPS_LABEL), renamer(BE_NUMBER_OF_CORES_LABEL)],
            ).style.applymap(
                partial(composite_latency_colors, slo=self.slo),
            ).format(
                partial(composite_latency_formatter, normalized=normalized)
            ).set_caption(
                self._get_caption('latency[us]', normalized)
            )

    def filtered_df_pivot_ui(self,
                             rows=(SWAN_AGGRESSOR_NAME_LABEL, 'be_l3_cache_ways'),
                             cols=(SWAN_LOAD_POINT_QPS_LABEL, 'be_number_of_cores'),
                             aggregatorName='First', vals=('percentile/99th',), rendererName='Heatmap', **options):
        return _pivot_ui(
            self.filtered_df(),
            totals=False,
            rows=rows,
            cols=cols,
            vals=vals,
            aggregatorName=aggregatorName,
            rendererName=rendererName,
            **options
        )
