import csv
import json

class Metric:
    def __init__(self):
        self.ns = None
        self.ver = None
        self.host = None
        self.time = None
        self.boolval = None
        self.doubleval = None
        self.strval = None
        self.tags = None
        self.valtype = None

    def __repr__(self):
        return "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s" % (self.ns, self.ver, self.host, self.time, self.boolval, self.doubleval, self.strval, self.tags, self.valtype)

def strip_quotation(input):
    output = input

    if len(input) < 2:
        return input

    if output[0] == '\'':
        output = output[1:]

    if output[-1] == '\'':
        output = output[:-1]

    return output

def is_null(input):
    if input == "null":
        return None

    return input

def read(path):
    output = []
    field_to_name = {}
    with open(path, 'rb') as csvfile:
        metric_reader = csv.reader(csvfile, delimiter=',', quotechar='\'')
        for row in metric_reader:
            metric = Metric()
            
            if row[0] == "#ns":
                continue

            if len(row) != 9:
                continue

            metric.ns = is_null(row[0])
            metric.ver = int(row[1])
            metric.host = is_null(strip_quotation(row[2]))
            metric.time = is_null(strip_quotation(row[3]))
            metric.boolval = bool(is_null(row[4]))
            metric.doubleval = float(is_null(row[5]))
            metric.strval = is_null(strip_quotation(row[6]))
            metric.tags = json.loads(is_null(strip_quotation(row[7])))
            metric.valtype = is_null(strip_quotation(row[8]))

            output.append(metric)

    return output
