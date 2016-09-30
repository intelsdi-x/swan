"""
This module contains the logic to render a sensivity profile (table) for samples in an Experiment.
"""
import plotly.graph_objs as go
import pandas as pd
import numpy as np

from IPython.core.display import HTML, display
from plotly.offline import init_notebook_mode, iplot

init_notebook_mode()


class Profile(object):
    """
    A sensivity profile is a table listing a workload's relative performance (it's measured
    quality metric against a target performance). The HTML representation of the profile color
    codes each cell based on it's slack (quality of service head room) or violation.
    """
    def __init__(self, e, slo):
        """
        :param e: an Experiment class object
        :param slo: performance target [int]

        Initializes a sensivity profile with `e` experiment object and visualize it against the
        specified slo (performance target).
        """
        self.categories = []
        self.data_frame = e.get_frame()
        self.data_visual = {}

        df_of_99th = self.data_frame.loc[self.data_frame['ns'].str.contains('/percentile/99th')]
        df_of_99th.is_copy = False
        df_of_99th['swan_aggressor_name'].replace(to_replace='None', value="Baseline", inplace=True)

        self.p99_by_aggressor = df_of_99th.groupby('swan_aggressor_name')
        columns = None
        data = []
        index = []

        for name, df in self.p99_by_aggressor:
            index.append(name)
            df.is_copy = False
            df.loc[:, 'swan_loadpoint_qps'] = pd.to_numeric(df['swan_loadpoint_qps'])
            aggressor_frame = df.sort_values('swan_loadpoint_qps')[['swan_loadpoint_qps', 'value']]
            self.categories.append(name)

            # Store columns for data frame from the target QPSes.
            # In case of partial measurements, we only use the columns from this aggressor
            # if it is bigger than the current one.
            qps = aggressor_frame['swan_loadpoint_qps'].tolist()
            if columns is None or len(columns) < len(qps):
                columns = qps

            violations = aggressor_frame['value'].apply(lambda x: (x / slo) * 100)
            data.append(violations.tolist())

            self.data_visual['x'] = np.array(qps)
            self.data_visual[name] = aggressor_frame['value'].as_matrix()
            self.data_visual['slo'] = [slo for i in qps]

        peak = np.amax(columns)
        columns = map(lambda c: (float(c) / peak) * 100, columns)
        self.frame = pd.DataFrame(data, columns=columns, index=index).sort_index()

    def sensitivity_table(self):
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

        return HTML(html_out)

    def sensitivity_chart(self):
        df = pd.DataFrame.from_dict(self.data_visual, orient='index').T
        data = list()
        data.append(go.Scatter(
            x=df['x'],
            y=df['slo'],
            fill=None,
            name='slo',
            mode='lines',
            line=dict(
                color='rgb(255, 0, 0)',
            )
        ))
        for agr in self.categories:
            data.append(go.Scatter(
                    x=df['x'],
                    y=df[agr],
                    name=agr,
                    fill='tozeroy',
                    mode='lines',
                ))

        display(iplot(data))


if __name__ == '__main__':
    from experiment import Experiment

    exp = Experiment(experiment_id='ad8b76f5-e627-4e9a-53b3-0b20117b7394', cassandra_cluster=['127.0.0.1'], port=19042)
    Profile(exp, slo=500).sensitivity_table()
    Profile(exp, slo=500).sensitivity_chart()

