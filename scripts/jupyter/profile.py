"""
This module contains the logic to render a sensivity profile (table) for samples in an Experiment.
"""

import collections

class Profile(object):
    """
    A sensivity profile is a table listing a workload's relative performance (it's measured
    quality metric against a target performance). The HTML representation of the profile color
    codes each cell based on it's slack (quality of service head room) or violation.
    """

    def __init__(self, samples, slo):
        """
        Initializes a sensivity profile with given list of Sample objects and visualized against the
        specified slo (performance target).
        """
        # TODO: slo should be read from database.
        self.slo = slo
        self.sensivity_rows = collections.OrderedDict()
        self.phase_to_aggressor = {}

        for sample_row in samples:
            # Categorize for sensitivity profile.
            if 'swan_loadpoint_qps' in sample_row.tags:
                phase = sample_row.tags['swan_phase']

                # Ugly hack. Phase names are repeated and specialized from Swan side.
                # Should be fixed when we introduce configuration unfolding.
                phase = phase.split('_id_')[0]

                if 'swan_aggressor_name' in sample_row.tags:
                    self.phase_to_aggressor[phase] = sample_row.tags['swan_aggressor_name']

                load_point = int(sample_row.tags['swan_loadpoint_qps'])

                if phase not in self.sensivity_rows:
                    self.sensivity_rows[phase] = collections.OrderedDict()

                if load_point not in self.sensivity_rows[phase]:
                    self.sensivity_rows[phase][load_point] = {}

                self.sensivity_rows[phase][load_point][sample_row.metric_name()] = sample_row

    def _repr_html_(self):
        # HTML styling constants
        no_border = 'border: 0'
        black_border = '1px solid black'

        html_out = ''
        html_out += '<table style="border: 0;">'
        html_out += '<tr style="%s">' % no_border
        html_out += '<th style="%s; border-bottom: %s; border-right: %s;">Scenario / Load</th>' % \
            (no_border, black_border, black_border)

        for load_percentage in range(5, 100, 10):
            html_out += '<th style="border: 0; border-bottom: 1px solid black;">%s%%</th>' % \
                load_percentage

        html_out += '</tr>'

        for phase in self.sensivity_rows:
            html_out += '<tr style="%s">' % no_border

            aggressor = self.phase_to_aggressor[phase]
            if aggressor == 'None':
                label = 'Baseline'
            else:
                label = aggressor

            html_out += '<td style="%s; border-right: %s;">%s</td>' % \
                (no_border, black_border, label)

            for load_points in self.sensivity_rows[phase]:
                samples = self.sensivity_rows[phase][load_points]

                if 'percentile/99th' in samples:
                    latency = samples['percentile/99th']
                    violation = ((latency.doubleval / self.slo) * 100)
                    style = '%s; ' % no_border

                    if violation > 150:
                        style += 'background-color: #a9341f; color: white;'
                    elif violation > 100:
                        style += 'background-color: #ffeda0;'
                    else:
                        style += 'background-color: #98cc70;'

                    html_out += '<td style="%s">%.1f%%</td>' % (style, violation)
                else:
                    html_out += '<td style="%s"></td>' % no_border
                    html_out += '</tr>'

        html_out += '</table>'

        return html_out
