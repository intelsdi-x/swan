import unittest

import experiment


class TestExperiment(unittest.TestCase):
    def test_zero_samples(self):
        exp = experiment.Experiment(experiment_id='', read_csv='test_data/empty.csv')
        self.assertTrue(len(exp.frame.index) == 0)

    def test_several_samples(self):
        exp = experiment.Experiment(
            experiment_id='8ab3f479-a3f8-48cf-71cb-e4853caf9cac',
            read_csv='test_data/experiments.csv')
        self.assertTrue(len(exp.frame.index) == 9)

if __name__ == '__main__':
    unittest.main()
