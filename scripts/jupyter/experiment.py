"""
This module contains the convience class to read experiment data and generate sensitivity
profiles. See profile.py for more information.
"""

from cassandra.cluster import Cluster
from cassandra.query import SimpleStatement
from IPython.core.display import display, HTML, clear_output
import numpy as np
import pandas as pd
import test_data_reader

class Experiment(object):
    """
    The Experiment class works as a container for swan experiment metrics.
    The experiment id should either be found when running the experiment through the swan cli or
    from using the Experiments class.
    """

    def __init__(self, experiment_id, **kwargs):
        """
        Initializes an experiment from a given experiment id by using the cassandra cluster and
        session.
        Set cassandra_cluster to an array of hostnames where cassandra nodes reside.
        """

        self.experiment_id = experiment_id
        self.phases = {}

        tag_column_map = {}
        columns = []
        data = []

        def add_sample(row):
            if row.valtype == "doubleval":
                values = [row.ns, row.host, row.time, row.doubleval]

                tag_values = [None] * len(tag_column_map)
                for key, value in row.tags.iteritems():
                    index = 0
                    if key not in tag_column_map:
                        # Find a position for in the array
                        index = len(tag_column_map)
                        tag_column_map[key] = index
                    else:
                        index = tag_column_map[key]

                    if len(tag_values) <= index:
                        tag_values.append(None)

                    # See if value can be converted to float
                    try:
                        tag_values[index] = float(value)
                    except ValueError:
                        tag_values[index] = value

                values.extend(tag_values)
                data.append(values)


        if 'cassandra_cluster' in kwargs:
            self.cluster = Cluster(kwargs['cassandra_cluster'])
            self.session = self.cluster.connect('snap')

            query = 'SELECT ns, ver, host, time, boolval, doubleval, strval, tags, valtype \
                 FROM snap.metrics WHERE tags CONTAINS \'%s\' ALLOW FILTERING' % self.experiment_id
            statement = SimpleStatement(query, fetch_size=100)

            # Iterator for the 'progress' bar.
            row_count = 1
            for row in self.session.execute(statement):
                # Temporary 'progress' bar in viewer to see progress of loading data.
                clear_output(wait=True)
                display(HTML('Loading row %d' % row_count))
                row_count += 1

                add_sample(row)

            # Clear output so the last number of loaded rows doesn't show.
            clear_output(wait=False)

        if 'test_file' in kwargs:
            test_rows = test_data_reader.read(kwargs['test_file'])
            for test_row in test_rows:
                add_sample(test_row)


        columns = ['ns', 'host', 'time', 'value']
        columns.extend([None] * len(tag_column_map))
        for tag, index in tag_column_map.iteritems():
            columns[4 + index] = tag
        self.frame = pd.DataFrame(data, columns=columns)

    def _repr_html_(self):
        return self.frame.to_html(index=False)
