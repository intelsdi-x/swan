"""
 Copyright (c) 2017 Intel Corporation

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
"""

import unittest

import experiment


class TestExperiment(unittest.TestCase):
    def test_not_existing_samples(self):
        with self.assertRaises(IOError) as context:
            experiment.Experiment(experiment_id='not existing experiment', name='first exp',
                                  read_csv=True, dir_csv='test_data')

        self.assertTrue(IOError, 'File test_data/not existing experiment.csv does not exist' in context.exception)

    def test_empty_sample(self):
        exp = experiment.Experiment(experiment_id='empty', read_csv=True, dir_csv='test_data')
        self.assertTrue(len(exp.frame.index) == 0)

    def test_several_samples(self):
        exp = experiment.Experiment(experiment_id='afa8531c-ab17-4a76-696c-90a14c0bda07',
                                    read_csv=True,  dir_csv='test_data')
        self.assertTrue(len(exp.frame.index) == 180)


if __name__ == '__main__':
    unittest.main()
