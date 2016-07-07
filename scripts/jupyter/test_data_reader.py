"""
This module contains a helper to generate list of Sample objects from a comma separated file.
"""

import csv
import json
from sample import Sample

def strip_quotation(input):
    """
    The test metrics files adds quotations to string fields for readability.
    For example:
    '/intel/swan/mutilate/machine1/std',-1,'localhost.localdomain',...
    The incoming field may therefore include quotation, which needs to be removed before we store
    the sample.
    """

    if input is None:
        return input

    output = input.strip()

    if len(input) < 2:
        return input

    if output[0] == '\'':
        output = output[1:]

    if output[-1] == '\'':
        output = output[:-1]

    return output

def is_null(input):
    """
    The test metrics files need to have a way to express that a value is absent.
    This is done by setting the value to 'null'. is_null converts this to a None type.
    """

    if input == 'null':
        return None

    return input

def read(path):
    """
    read returns a list of Samples
    """

    output = []
    field_to_name = {}
    with open(path, 'rb') as csvfile:
        sample_reader = csv.reader(csvfile, delimiter=',', quotechar='\'')
        for row in sample_reader:
            sample = Sample()

            if row[0] == '#ns':
                continue

            if len(row) != 9:
                continue

            sample.ns = is_null(row[0])
            sample.ver = int(row[1])
            sample.host = is_null(strip_quotation(row[2]))
            sample.time = is_null(strip_quotation(row[3]))
            sample.boolval = bool(is_null(row[4]))
            sample.doubleval = float(is_null(row[5]))
            sample.strval = is_null(strip_quotation(row[6]))
            sample.tags = json.loads(is_null(strip_quotation(row[7])))
            sample.valtype = is_null(strip_quotation(row[8]))

            output.append(sample)

    return output
