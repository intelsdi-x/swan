"""
This module contains a helper to generate list of Sample objects from a comma separated file.
"""
import collections
import csv
import json


def strip_quotation(input):
    """
    The test metrics files adds quotations to string fields for readability.
    For example:
    '/intel/swan/mutilate/machine1/std',-1,'localhost.localdomain',...
    The incoming field may therefore include quotation, which needs to be removed before we store
    the sample.
    """

    if input:
        output = input.strip()
        if len(input) < 2:
            return input
        if output.startswith('\''):
            output = output[1:]
        if output.endswith('\''):
            output = output[:-1]

    return output


def read(path):
    """
    read returns a list of Samples
    """

    rows = {}
    qps = {}

    convert_null = lambda i: None if i == 'null' else i
    Sample = collections.namedtuple('Sample', 'ns ver host time boolval doubleval strval tags valtype')

    with open(path, 'rb') as csvfile:
        sample_reader = csv.reader(csvfile, delimiter=',', quotechar='\'')
        for row in sample_reader:
            if row[0] == '#ns':
                continue

            if len(row) != 9:
                continue

            sample = Sample(ns=convert_null(row[0]), ver=int(row[1]), host=convert_null(strip_quotation(row[2])),
                            time=convert_null(strip_quotation(row[3])), boolval=bool(convert_null(row[4])),
                            doubleval=float(convert_null(row[5])), strval=convert_null(strip_quotation(row[6])),
                            tags=json.loads(convert_null(strip_quotation(row[7]))),
                            valtype=convert_null(strip_quotation(row[8])))

            if "/intel/swan/mutilate/%s/qps" % sample.host == row[0]:
                qps[(sample.tags['swan_aggressor_name'], sample.tags['swan_phase'],
                     sample.tags['swan_repetition'])] = sample.doubleval
            k = (sample.ns, sample.tags['swan_aggressor_name'], sample.tags['swan_phase'],
                 sample.tags['swan_repetition'])
            rows[k] = sample

    return rows, qps
