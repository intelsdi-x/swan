"""
This module contains the logic to render a sensivity profile (table) for samples in an Experiment.
"""

import pandas as pd

class Profile(object):
    """
    A sensivity profile is a table listing a workload's relative performance (it's measured
    quality metric against a target performance). The HTML representation of the profile color
    codes each cell based on it's slack (quality of service head room) or violation.
    """

    def __init__(self, data_frame, slo):
        """
        Initializes a sensivity profile with given list of Sample objects and visualized against the
        specified slo (performance target).
        """

        p99  = data_frame.loc[data_frame['ns'].str.contains('/percentile/99th')]
        p99_by_aggressor = p99.groupby('swan_aggressor_name')
        columns = None
        profile = None
        data = []
        index = []

        def percentage_of_slo(x):
            return (x / slo) * 100

        for name, df in p99_by_aggressor:
            # Overwrite the 'None' aggressor with 'Baseline'
            if name == 'None':
                name = 'Baseline'

            index.append(name)

            aggressor_frame = df.sort_values('swan_loadpoint_qps')[['swan_loadpoint_qps', 'value']]

            violations = aggressor_frame['value'].apply(percentage_of_slo)

            # Store columns for data frame from the target QPSes.
            # In case of partial measurements, we only use the columns from this aggressor
            # if it is bigger than the current one.
            qps = aggressor_frame['swan_loadpoint_qps'].tolist()
            if columns is None:
                columns = qps
            elif len(columns) < len(qps):
                columns = qps

            data.append(violations.tolist())

        if columns is not None:
            # Apply filter to columns to enable other formatting.
            peak = max(columns)
            def percentage_of_peak(qps):
                return (qps / peak) * 100
            columns = map(percentage_of_peak, columns)

            profile = pd.DataFrame(data, columns=columns, index=index).sort_index()


        self.frame = profile

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
                else:
                    style += 'background-color: #98cc70;'

                html_out += '<td style="%s">%.1f%%</td>' % (style, value)

            html_out += '</tr>'

        html_out += '</table>'

        return html_out
