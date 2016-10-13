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
        self.exp = e
        self.slo = slo
        self.categories = []
        self.data_frame = self.exp.get_frame()
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

        self.data_visual_frame = pd.DataFrame.from_dict(self.data_visual, orient='index').T

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

    def sensitivity_chart(self, fill=False, to_max=False):
        categories = self.categories
        gradients = ['rgb(7, 249, 128)', 'rgb(127, 255, 158)', 'rgb(243, 255, 8)', 'rgb(255, 178, 54)',
                     'rgb(255, 93, 13)', 'rgb(255, 31, 10)', 'rgb(255, 8, 0)']
        data = list()
        fill_to_nexty = 'tonexty' if fill else None
        data.append(go.Scatter(
            x=self.data_visual_frame['x'],
            y=self.data_visual_frame['slo'],
            fill=None,
            name='slo',
            mode='lines',
            line=dict(
                color='rgb(255, 0, 0)',
                shape='spline'
            )
        ))

        if to_max:
            self.data_visual_frame['max_aggrs'] = self.data_visual_frame[
                ['L1 Instruction', 'Baseline', 'Stream 100M', 'Caffe', 'L1 Data', 'L3 Data']].max(axis=1)
            categories = ['Baseline', 'max_aggrs']

        for i, agr in enumerate(categories):
            scatter_aggr = go.Scatter(
                x=self.data_visual_frame['x'],
                y=self.data_visual_frame[agr],
                name=self.exp.name + ':' + agr,
                fill=fill_to_nexty if agr != 'Baseline' else None,
                mode='lines',
                line=dict(
                    shape='spline',
                ),
            )
            if fill:
                scatter_aggr['line']['color'] = gradients[i]

            data.append(scatter_aggr)

        layout = go.Layout(
            xaxis=dict(
                title='QPS',
                titlefont=dict(
                    family='Arial, sans-serif',
                    size=18,
                    color='lightgrey'
                ),
            ),
            yaxis=dict(
                range=[0, 2 * self.slo],
                title='Latency',
                titlefont=dict(
                    family='Arial, sans-serif',
                    size=18,
                    color='lightgrey'
                ),
            )
        )

        fig = go.Figure(data=data, layout=layout)
        display(iplot(fig))


def compare_experiments(exps, slo=500, aggressors=None, fill=False, to_max=False):
    if not aggressors:
        aggressors = ["Baseline", ]

    data = []
    for exp in exps:
        df = Profile(exp, slo).data_visual_frame

        if to_max:
            df['max_aggrs'] = df[['L1 Instruction', 'Baseline', 'Stream 100M',
                                  'Caffe', 'L1 Data', 'L3 Data']].max(axis=1)
            aggressors = ["Baseline", 'max_aggrs']

        for aggressor in aggressors:
            data.append(
                go.Scatter(
                    x=df['x'],
                    y=df[aggressor],
                    fill=None,
                    name='%s:%s' % (exp.name, aggressor),
                    mode='lines',
                    line=dict(
                        shape='spline'
                    )
                )
            )

    max_x_series = np.maximum(data[0]['x'], data[2]['x'])
    all_set = np.append(data[0]['x'], data[2]['x'])
    max_x_series[0] = np.min(all_set)
    max_x_series[-1] = np.max(all_set)

    slo_scatter = go.Scatter(
        x=max_x_series,
        y=[slo for i in max_x_series],
        fill=None,
        name="slo",
        mode='lines',
        line=dict(
            color='rgb(255, 0, 0)',
        )
    )
    data.append(slo_scatter)

    layout = go.Layout(
        xaxis=dict(
            title='QPS',
            titlefont=dict(
                family='Arial, sans-serif',
                size=18,
                color='lightgrey'
            ),
        ),
        yaxis=dict(
            range=[0, 2 * slo],
            title='Latency',
            titlefont=dict(
                family='Arial, sans-serif',
                size=18,
                color='lightgrey',
            ),
        )
    )

    # Fill only if there is two experiments with one aggressor or one experiment with two aggressors to compare
    if fill and ((len(exps) == 2 and len(aggressors) == 1) or (len(exps) == 1 and len(aggressors) == 2)):
        data[0]['fill'] = None
        data[1]['line']['color'] = 'rgb(7, 249, 128)'
        data[1]['fill'] = 'tonexty'

    fig = go.Figure(data=data, layout=layout)
    display(iplot(fig))


if __name__ == '__main__':
    from experiment import Experiment

    exp1 = Experiment(experiment_id='e527ce99-bcbc-4009-7367-b8234d0de89d', cassandra_cluster=['149.202.205.137'], port=9042,
                      name='exp_with_serenity')
    exp2 = Experiment(experiment_id='dfdca05e-881a-4416-5e8f-37cf0fd26827', cassandra_cluster=['149.202.205.137'], port=9042,
                      name='exp_without_serenity')

    Profile(exp1, slo=500).sensitivity_table()
    Profile(exp2, slo=500).sensitivity_chart()

    compare_experiments(exps=[exp1, exp2], aggressors=["Baseline"], fill=True, to_max=True)
