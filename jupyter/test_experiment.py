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
