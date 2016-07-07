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
