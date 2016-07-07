import unittest
import experiment

class TestExperiment(unittest.TestCase):
    def test_zero_samples(self):
        exp = experiment.Experiment(experiment_id='', test_file='test_data/empty.csv')
        self.assertTrue(len(exp.samples) == 0)

    def test_several_samples(self):
        exp = experiment.Experiment(experiment_id='8ab3f479-a3f8-48cf-71cb-e4853caf9cac', test_file='test_data/one_experiment.csv')
        self.assertTrue(len(exp.samples) == 9)


if __name__ == '__main__':
    unittest.main()
