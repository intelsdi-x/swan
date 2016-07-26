"""
This module contains the logic to render a sensivity profile (table) for samples in an Experiment.
"""
import pandas as pd
import numpy as np


class Profile(object):
    """
    A sensivity profile is a table listing a workload's relative performance (it's measured
    quality metric against a target performance). The HTML representation of the profile color
    codes each cell based on it's slack (quality of service head room) or violation.
    """
    def __init__(self, e, slo):
        """
        Initializes a sensivity profile with given list of Sample objects and visualized against the
        specified slo (performance target).
        """

        data_frame = e.get_frame()
        df_of_99th = data_frame.loc[data_frame['ns'].str.contains('/percentile/99th')]
        df_of_99th.is_copy = False
        df_of_99th['swan_aggressor_name'].replace(to_replace='None', value="Baseline", inplace=True)

        p99_by_aggressor = df_of_99th.groupby('swan_aggressor_name')
        columns = None
        data = []
        index = []

        for name, df in p99_by_aggressor:
            index.append(name)
            df.is_copy = False
            df.loc[:, 'swan_loadpoint_qps'] = pd.to_numeric(df['swan_loadpoint_qps'])
            aggressor_frame = df.sort_values('swan_loadpoint_qps')[['swan_loadpoint_qps', 'value']]

            # Store columns for data frame from the target QPSes.
            # In case of partial measurements, we only use the columns from this aggressor
            # if it is bigger than the current one.
            qps = aggressor_frame['swan_loadpoint_qps'].tolist()
            if columns is None or len(columns) < len(qps):
                columns = qps

            violations = aggressor_frame['value'].apply(lambda x: (x / slo) * 100)
            data.append(violations.tolist())

        peak = np.amax(columns)
        columns = map(lambda c: (float(c) / peak) * 100, columns)
        self.frame = pd.DataFrame(data, columns=columns, index=index).sort_index()

    def _repr_html_(self):
        no_border = 'border: 0'
        black_border = '1px solid black'
        html_out = ''
        html_out += '<table style="%s">' % no_border
        html_out += '<tr style="%s">' % no_border
        html_out += '<th style="%s; border-bottom: %s; border-right: %s;">Scenario / Load</th>' % \
            (no_border, black_border, black_border)

        for column in self.frame:
            html_out += '<th style="border: 0; border-bottom: 1px solid black;">%s%%</th>' % \
                column

        html_out += '</tr>'

        for index, row in self.frame.iterrows():
            html_out += '<tr style="%s">' % no_border
            aggressor = index
            if aggressor == 'None':
                label = 'Baseline'
            else:
                label = aggressor

            html_out += '<td style="%s; border-right: %s;">%s</td>' % \
                (no_border, black_border, label)

            for value in row:
                style = '%s; ' % no_border
                if value > 150:
                    style += 'background-color: #a9341f; color: white;'
                elif value > 100:
                    style += 'background-color: #ffeda0;'
                elif np.isnan(value):
                    value = 0
                    style += 'background-color: #a9341f; color: white;'
                else:
                    style += 'background-color: #98cc70;'

                html_out += '<td style="%s">%.1f%%</td>' % (style, value)

            html_out += '</tr>'

        html_out += '</table>'

        return html_out


if __name__ == '__main__':
    from experiment import Experiment

    exp = Experiment(experiment_id='57d25f69-d6d7-43e1-5c4e-3b5f5208acdc', cassandra_cluster=['127.0.0.1'], port=9042)
    p = Profile(exp, slo=500)
