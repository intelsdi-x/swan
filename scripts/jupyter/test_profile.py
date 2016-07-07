import unittest
import experiment

class TestSensitivityProfile(unittest.TestCase):
    def test_empty_profile(self):
        exp = experiment.Experiment(experiment_id='8ab3f479-a3f8-48cf-71cb-e4853caf9cac', test_file='test_data/empty.csv')
        self.assertTrue(len(exp.samples) == 0)
        profile = exp.profile(500)
        self.assertTrue(len(profile.sensivity_rows) == 0)

    def test_single_field_profile(self):
        exp = experiment.Experiment(experiment_id='8ab3f479-a3f8-48cf-71cb-e4853caf9cac', test_file='test_data/one_experiment.csv')
        self.assertTrue(len(exp.samples) == 9)
        profile = exp.profile(500)
        self.assertTrue(len(profile.sensivity_rows) == 1)
