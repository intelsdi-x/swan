import unittest
import experiments

class TestExperiments(unittest.TestCase):
    def test_zero_experiments(self):
        exp = experiments.Experiments(test_file='test_data/empty.csv')
        self.assertTrue(len(exp.experiments) == 0)

    def test_one_experiment(self):
        exp = experiments.Experiments(test_file='test_data/one_experiment.csv')
        self.assertTrue(len(exp.experiments) == 1)
        self.assertTrue('8ab3f479-a3f8-48cf-71cb-e4853caf9cac' in exp.experiments)

    def test_several_experiments(self):
        exp = experiments.Experiments(test_file='test_data/experiments.csv')
        self.assertTrue(len(exp.experiments) == 3)
        self.assertTrue('8ab3f479-a3f8-48cf-71cb-e4853caf9cac' in exp.experiments)
        self.assertTrue('be996d77-4cf2-406a-a763-4b9d81b46f15' in exp.experiments)
        self.assertTrue('c98e53c4-d660-4a97-9713-97462438133f' in exp.experiments)


if __name__ == '__main__':
    unittest.main()
