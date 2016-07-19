"""
This module contains a helper to generate list of Sample objects from a comma separated file.
"""

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

def convert_null(input):
    """
    The test metrics files need to have a way to express that a value is absent.
    This is done by setting the value to 'null'. convert_null converts this to a None type.
    """

    if input == 'null':
        return None

    return input

def read(path):
    """
    read returns a list of Samples
    """

    output = []
    with open(path, 'rb') as csvfile:
        sample_reader = csv.reader(csvfile, delimiter=',', quotechar='\'')
        for row in sample_reader:
            if row[0] == '#ns':
                continue

            if len(row) != 9:
                continue

            class Sample:
                pass

            sample = Sample()

            setattr(sample, 'ns', convert_null(row[0]))
            setattr(sample, 'ver', int(row[1]))
            setattr(sample, 'host', convert_null(strip_quotation(row[2])))
            setattr(sample, 'time', convert_null(strip_quotation(row[3])))
            setattr(sample, 'boolval', bool(convert_null(row[4])))
            setattr(sample, 'doubleval', float(convert_null(row[5])))
            setattr(sample, 'strval', convert_null(strip_quotation(row[6])))
            setattr(sample, 'tags', json.loads(convert_null(strip_quotation(row[7]))))
            setattr(sample, 'valtype', convert_null(strip_quotation(row[8])))

            output.append(sample)

    return output
