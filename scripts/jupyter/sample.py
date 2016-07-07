import json

class Sample:
    def __init__(self, ns, ver, host, time, boolval, doubleval, strval, tags, valtype):
        self.ns = ns
        self.ver = ver
        self.host = host
        self.time = time
        self.boolval = boolval
        self.doubleval = doubleval
        self.strval = strval
        self.tags = tags
        self.valtype = valtype

    def metric_name(self):
        # Sanitize metric name by stripping namespace prefix.
        # For example, from '/intel/swan/mutilate/jp7-1/percentile/99th' to
        # 'percentile/99th'.

        # Make sure it is a metric we recognise.
        if not self.ns.startswith('/intel/swan/mutilate/'):
            return self.ns

        namespace_exploded = self.ns.split('/')

        # Make sure we have at least '/intel/swan/mutilate/', the host name and a metric.
        if len(namespace_exploded) < 5:
            return self.ns

        metric_exploded = namespace_exploded[5:]
        return '/'.join(metric_exploded)

    def _repr_html_(self):
        html_out = ""
        html_out += "<table>"
        html_out += "<tr><th>Namespace</th><th>Version</th><th>Host</th><th>Time</th>\
            <th>Value</th><th>Tags</th></tr>"

        value = 0.0

        html_out += "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%f</td>\
            <td><code>%s</code></td><tr>" % \
            (self.ns,
             self.ver,
             self.host,
             self.time,
             value,
             json.dumps(self.tags, sort_keys=True, indent=4, separators=(',', ': ')))

        html_out += '</table>'

        return html_out

    def __repr__(self):
        return json.dumps({
            'ns': self.ns,
            'ver': self.ver,
            'host': self.host,
            'time': self.time,
            'boolval': self.boolval,
            'doubleval': self.doubleval,
            'strval': self.strval,
            'tags': self.tags,
            'valtype': self.valtype
        })
