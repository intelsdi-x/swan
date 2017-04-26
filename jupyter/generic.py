import sys
import re
from functools import partial

import pandas as pd
from IPython.display import display, HTML

CASSANDRA_SESSION = None  # one instance for all existing notebook experiments
DEFAULT_CASSANDRA_OPTIONS = dict(
    cassandra_cluster=['127.0.0.1'],
    port=9042,
    ssl_options=None
)


def _create_or_get_session(cassandra_cluster, port, ssl_options):
    """ Get or preapare new session to Cassandra cluster """
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

        cluster = Cluster(cassandra_cluster, port=port, ssl_options=ssl_options, auth_provider=auth_provider)
        CASSANDRA_SESSION = cluster.connect()
        CASSANDRA_SESSION.row_factory = ordered_dict_factory
    return CASSANDRA_SESSION


pattern = re.compile(r'(/intel/swan/(\w+)/(\w+)/).*?')


def drop_prefix(value):
    """ Drop prefix /intel/swan/PLUGIN/HOST from ns column."""
    return pattern.sub(lambda x: '', value)


def load_from_cassandra(experiment_id, session, keyspace='snap'):
    """ Load data from cassandra database as dataframe.
    """
    from cassandra.query import SimpleStatement

    query = """SELECT ns, ver, host, time, boolval, doubleval, strval, tags, valtype
               FROM %s.metrics
               WHERE tags['swan_experiment'] = \'%s\'  ALLOW FILTERING""" % (
                   keyspace, experiment_id)

    statement = SimpleStatement(query)
    rows = list(session.execute(statement))
    if len(rows) == 0:
        print >>sys.stderr, "no metrics found!"
        return []
    return rows


def convert_rows_to_dataframe(rows):
    """ Convert rows (dicts) representing snap.metrics to pandas.DataFrame."""
    tags_keys = None

    # flatMap tags into new columns.
    for row in rows:
        if tags_keys is None:
            tags_keys = row.get('tags').keys()
        row.update(row.pop('tags'))

    # onvert simple Row objects (composite OrderedDicts) into dataframe.
    df = pd.DataFrame.from_records(rows)

    # Drop "/intel/swan/PLUGIN/HOST/" from ns.
    df['ns'] = df['ns'].apply(drop_prefix)

    # Convert all series to numeric if possible.
    for column in df.columns:
        try:
            df[column] = df[column].apply(pd.to_numeric)
        except ValueError:
            continue

    # Reshape - to have have all tags + "ns" column as index.

    # First tags and ns column is converted to index with group by.
    grouper = df.groupby(tags_keys+['ns'])
    # and take just an value of metric.
    grouper = grouper['doubleval']
    # Second - assuming count = 1, one can get a mx
    df = grouper.first()
    # Then use 'ns' categorical column values as new columns and drop index.
    df = df.unstack(['ns']).reset_index()

    return df


##################################
# Style functions for signle cells
##################################
CRIT = 'background:#a9341f; color: white;'
WARN = 'background:#ffeda0'
OK = 'background:#98cc70'
NAN = 'background-color: #c0c0c0'


def latency_colors(latency, slo):
    """ Style function for latency colors. """
    if pd.isnull(latency):
        return NAN
    if latency > 1.5 * slo:
        return CRIT
    elif latency > slo:
        return WARN
    else:
        return OK


def qps_colors(qps):
    """ Style function for qps colors. """
    if pd.isnull(qps):
        return NAN
    if qps > 0.95:
        return OK
    elif qps > 0.8:
        return WARN
    else:
        return CRIT

achieved_qps_label = 'achieved QPS'
both_label = 'both'


def _load_data_and_rename(experiment_id, cassandra_options):
    session = _create_or_get_session(**(cassandra_options or DEFAULT_CASSANDRA_OPTIONS))
    rows = load_from_cassandra(experiment_id, session)
    df = convert_rows_to_dataframe(rows)

    # Extra columns.
    # Calculate achieved QPS as percentage (normalized to 1).
    df[achieved_qps_label] = pd.Series(df['qps'] / df['swan_loadpoint_qps'])

    # Composite value to store both "achieved qps" and "latency" together as dict.
    columns = [achieved_qps_label, 'percentile/99th']
    df[both_label] = df[columns].apply(dict, axis=1)

    columns_to_rename = {
        'number_of_cores': 'Number of cores',
        'swan_loadpoint_qps': 'Target QPS',
        'percentile/99th': 'Tail latency (99th percentile) [us]',
    }

    def _renamed(original_name):
        return columns_to_rename.get(original_name, original_name)

    df = df.rename(columns=columns_to_rename)

    return df, _renamed


def optimal_core_allocation_latency(experiment_id, slo, cassandra_options=None):
    df, _renamed = _load_data_and_rename(experiment_id, cassandra_options)
    # latency table
    return df.pivot_table(
            values=_renamed('percentile/99th'),
            index=_renamed('number_of_cores'),
            columns=_renamed('swan_loadpoint_qps'),

        ).style.applymap(
            partial(latency_colors, slo=slo)
        ).format('{:.0f}')


def optimal_core_allocation_qps(experiment_id, slo, cassandra_options=None):
    """ Generate QPS tables """
    df, _renamed = _load_data_and_rename(experiment_id, cassandra_options)

    return df.pivot_table(
            values=achieved_qps_label,
            index=_renamed('number_of_cores'),
            columns=_renamed('swan_loadpoint_qps'),
        ).style.applymap(
            qps_colors
        ).format('{:.0%}')


def optimal_core_allocation(experiment_id, slo, cassandra_options=None):
    """ Generate raport """
    df, _renamed = _load_data_and_rename(experiment_id, cassandra_options)

    def formatter(d):
        if d is None:
            return 'N/A'
        qps = d[achieved_qps_label]
        latency = d['percentile/99th']
        if qps < 0.9:
            return 'FAIL'
        return '{:.0f}'.format(latency)

    def styler(d):
        if pd.isnull(d):
            return NAN

        qps = d[achieved_qps_label]
        latency = d['percentile/99th']
        if pd.isnull(qps) or pd.isnull(latency):
            return NAN

        if qps < 0.9:
            return NAN

        if latency > 1.5 * slo:
            return CRIT
        elif latency > slo:
            return WARN
        else:
            return OK

    return df.pivot_table(
        values=both_label,
        index=_renamed('number_of_cores'),
        columns=_renamed('swan_loadpoint_qps'),
        aggfunc='first',
    ).style.applymap(styler).format(formatter)


def pivot_ui(experiment_id, cassandra_options=None):
    """ Helper function to start interactive pivot table."""
    try:
        from pivottablejs import pivot_ui
        session = _create_or_get_session(**(cassandra_options or DEFAULT_CASSANDRA_OPTIONS))
        rows = load_from_cassandra(experiment_id, session)
        df = convert_rows_to_dataframe(rows)
        return pivot_ui(df)
    except ImportError:
        print >>sys.stderr, "Error: Cannot start interactive pivot table - please install pivottablejs egg!"
