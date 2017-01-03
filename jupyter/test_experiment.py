import unittest

import experiment


class TestExperiment(unittest.TestCase):
    def test_zero_samples(self):
        exp = experiment.Experiment(experiment_id='not existing experiment', name='first exp',
                                    read_csv=True)
        self.assertTrue(len(exp.frame.index) == 0)

    def test_several_samples(self):
        exp = experiment.Experiment(
            experiment_id='7be3c448-4fa2-4178-75aa-e23d292d4030',
            read_csv=True)
        self.assertTrue(len(exp.frame.index) == 1260)


if __name__ == '__main__':
    unittest.main()
